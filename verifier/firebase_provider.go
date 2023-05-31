package verifier

import (
	"crypto"
	"errors"
	"github.com/arcana-network/dkgnode/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
	"strconv"
	"strings"
	"time"
)

const (
	FirebaseKeyListURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

type FirebaseProvider struct {
	Leeway           time.Duration
	signingKeys      map[string]crypto.PublicKey
	signingKeyExpiry *time.Time
}

type FirebaseVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

func NewFirebaseProvider() (*FirebaseProvider, error) {
	f := &FirebaseProvider{
		Leeway: 120 * time.Second,
	}
	err := f.initializeKeys()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *FirebaseProvider) initializeKeys() error {
	// https://firebase.google.com/docs/auth/admin/verify-id-tokens#verify_id_tokens_using_a_third-party_jwt_library
	keyMap := make(map[string]string, 4)
	parsedKM := make(map[string]crypto.PublicKey, 4)

	rsp, err := req.R().SetSuccessResult(&keyMap).Get(FirebaseKeyListURL)
	if err != nil {
		return err
	}

	var expiry *time.Time = nil
	ctrlHdr := rsp.Header.Get("Cache-Control")
	if len(ctrlHdr) != 0 {
		// This not a valid algorithm to parse the token, when the time comes, we should likely see use github.com/pquerna/cachecontrol/cacheobject's ParseResponseCacheControl.
		parts := strings.Split(ctrlHdr, ",")
		for _, part := range parts {
			sides := strings.Split(strings.TrimSpace(part), "=")
			if len(sides) != 2 {
				continue
			}
			if strings.ToLower(sides[0]) != "max-age" {
				continue
			}
			rhs, err := strconv.ParseUint(sides[1], 10, 32)
			if err != nil {
				return err
			}
			_exp := time.Now().Add(time.Duration(rhs) * time.Second)
			expiry = &_exp
			break
		}
	}

	for k, v := range keyMap {
		pk, err := jwt.ParseRSAPublicKeyFromPEM([]byte(v))
		if err != nil {
			return err
		}
		parsedKM[k] = pk
	}

	f.signingKeys = parsedKM
	f.signingKeyExpiry = expiry
	return nil
}

func (f *FirebaseProvider) ID() string {
	return "firebase"
}

func (f *FirebaseProvider) CleanToken(s string) string {
	return strings.Trim(s, " ")
}

func (f *FirebaseProvider) verifyJWT(token *jwt.Token) (interface{}, error) {
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
	k, ok := f.signingKeys[KIDStr]
	if !ok {
		return nil, errors.New("kid invalid (2)")
	}
	return k, nil

}

func (f *FirebaseProvider) Verify(message *bijson.RawMessage, params *common.VerifierParams) (verified bool, verifierID string, err error) {
	// Reinitialize the keys if they are expired already
	if f.signingKeyExpiry.Before(time.Now()) {
		err := f.initializeKeys()
		if err != nil {
			return false, "", err
		}
	}

	var p FirebaseVerifierParams
	if err := bijson.Unmarshal(*message, &p); err != nil {
		return false, "", err
	}
	tok, err := jwt.ParseWithClaims(p.IDToken, &jwt.RegisteredClaims{}, f.verifyJWT, jwt.WithAudience(params.ClientID), jwt.WithIssuer("https://securetoken.google.com/"+params.ClientID), jwt.WithIssuedAt(), jwt.WithLeeway(f.Leeway), jwt.WithValidMethods([]string{"RS256"}))
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
	if time.Now().Sub(iat.Time) > f.Leeway {
		return false, "", errors.New("token was issued way too long ago")
	}

	return true, subj, nil
}
