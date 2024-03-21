package verifier

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/imroc/req/v3"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type CustomProvider struct {
	Timeout time.Duration
}

type CustomProviderResponse struct {
	JWKUrl   string            `json:"jwkUrl"`
	IDParam  string            `json:"idParam"`
	Params   map[string]string `json:"params"`
	Issuer   string            `json:"issuer"`
	Audience string            `json:"audience"`
}

type CustomProviderParams struct {
	IDToken  string `json:"id_token"`
	UserID   string `json:"user_id"`
	AppID    string `json:"app_id"`
	Provider string `json:"provider"`
}

func NewCustomProvider() *CustomProvider {
	return &CustomProvider{
		Timeout: 60 * time.Second,
	}
}

func (t *CustomProvider) ID() string {
	return "custom"
}

func (t *CustomProvider) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (t *CustomProvider) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p CustomProviderParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	// Fetch creds from params
	u, err := url.Parse(config.GlobalConfig.GatewayURL)
	if err != nil {
		return false, "", err
	}
	u.Path = "/api/v2/provider"
	u.RawQuery = fmt.Sprintf("provider=%s&appID=%s", p.Provider, p.AppID)
	customProviderParams := CustomProviderResponse{}
	res, err := req.R().SetSuccessResult(&customProviderParams).Get(u.String())
	if err != nil {
		return false, "", err
	}

	if res.IsErrorState() {
		return false, "", errors.New("invalid provider")
	}

	log.WithFields(log.Fields{
		"url":      u.String(),
		"response": customProviderParams,
		"status":   res.StatusCode,
		"token":    p.IDToken,
	}).Debug("CustomProvider")
	// Fetch JWK from JWKUrl
	set, err := jwk.Fetch(context.Background(), customProviderParams.JWKUrl)
	if err != nil {
		log.WithError(err).WithField("url", customProviderParams.JWKUrl).Info("jwk.fetch: invalid jwkurl")
		return false, "", errors.New("invalid jwk url")
	}

	// Verify signature via JWK
	tok, err := jwt.Parse(
		[]byte(p.IDToken),
		jwt.WithKeySet(set, jws.WithRequireKid(false)),
		jwt.WithValidate(true),
	)
	if err != nil {
		log.WithError(err).Info("jwt.Parse")
		return false, "", fmt.Errorf("jwt parse error: %w", err)
	}

	if tok.IssuedAt().Before(time.Now().Add(-2 * time.Minute)) {
		return false, "", errors.New("jwt older than 2 minute")
	}

	id, ok := tok.Get(customProviderParams.IDParam)
	if !ok {
		log.WithError(err).Info("tok.Get[IDParam]")
		return false, "", errors.New("id is not set in provider")
	}
	idStr, ok := id.(string)
	if !ok {
		return false, "", errors.New("id not present in token")
	}
	if idStr != p.UserID {
		return false, "", errors.New("invalid user id")
	}
	if tok.Issuer() != customProviderParams.Issuer {
		return false, "", errors.New("invalid issuer")
	}
	if tok.Audience()[0] != customProviderParams.Audience {
		return false, "", errors.New("invalid audience")
	}

	// Verify according to data
	for k, v := range customProviderParams.Params {
		val, ok := tok.Get(k)
		if !ok {
			return false, "", fmt.Errorf("%s is not in token", k)
		}
		if val.(string) != v {
			return false, "", fmt.Errorf("%s not satisfied", k)
		}
	}

	return true, p.UserID, nil
}
