package verifier

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
)

const (
	AppleIDKeyListURL = "https://appleid.apple.com/auth/keys"
)

type AppleIDProvider struct {
	Leeway time.Duration
}

type AppleIDProviderParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

type AppleIDKeyList struct {
	Keys []struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		Use string `json:"use"`
		Alg string `json:"alg"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func NewAppleIDProvider() *AppleIDProvider {
	f := &AppleIDProvider{
		Leeway: 120 * time.Second,
	}
	return f
}

func (f *AppleIDProvider) getKeys() (map[string]crypto.PublicKey, error) {
	var response AppleIDKeyList
	parsedKM := make(map[string]crypto.PublicKey, 4)

	rsp, err := req.R().
		SetSuccessResult(&response).
		Get(AppleIDKeyListURL)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch the latest Apple ID keyset: %d", rsp.StatusCode)
	}

	for _, key := range response.Keys {
		nDecoded, err := base64.RawStdEncoding.DecodeString(key.N)
		if err != nil {
			return nil, err
		}
		nBig := new(big.Int).SetBytes(nDecoded)

		eDecoded, err := base64.RawStdEncoding.DecodeString(key.E)
		if err != nil {
			return nil, err
		}
		eReal := binary.BigEndian.Uint64(eDecoded)

		parsedKM[key.Kid] = &rsa.PublicKey{
			N: nBig,
			E: int(eReal),
		}
	}

	return parsedKM, nil
}

func (f *AppleIDProvider) ID() string {
	return "apple_id"
}

func (f *AppleIDProvider) CleanToken(s string) string {
	return strings.Trim(s, " ")
}

func (f *AppleIDProvider) Verify(message *bijson.RawMessage, params *common.VerifierParams) (verified bool, verifierID string, err error) {
	keys, err := f.getKeys()
	if err != nil {
		return false, "", err
	}

	var p AppleIDProviderParams
	if err := bijson.Unmarshal(*message, &p); err != nil {
		return false, "", err
	}

	tok, err := jwt.ParseWithClaims(
		p.IDToken,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != "RS256" {
				return nil, errors.New("invalid signing algorithm")
			}
			kid, ok := token.Header["kid"]
			if !ok {
				return nil, errors.New("kid missing")
			}
			KIDStr, ok := kid.(string)
			if !ok {
				return nil, errors.New("kid invalid (1)")
			}
			k, ok := keys[KIDStr]
			if !ok {
				return nil, errors.New("kid invalid (2)")
			}
			return k, nil
		},
		jwt.WithAudience(params.ClientID),
		// TODO – jwt.WithIssuer("apple"),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(f.Leeway),
		jwt.WithValidMethods([]string{"RS256"}))
	if err != nil {
		return false, "", err
	}
	if !tok.Valid {
		return false, "", errors.New("token not valid")
	}
	subj, err := tok.Claims.GetSubject()
	if err != nil {
		return false, "", err
	}
	iat, err := tok.Claims.GetIssuedAt()
	if err != nil {
		return false, "", err
	}
	if time.Since(iat.Time) > f.Leeway {
		return false, "", errors.New("token was issued way too long ago")
	}

	return true, subj, nil
}
