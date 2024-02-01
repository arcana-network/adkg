package verifier

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/torusresearch/bijson"
)

func NewTwitchProvider() *TwitchVerifier {
	provider, err := oidc.NewProvider(context.TODO(), "https://id.twitch.tv/oauth2")
	if err != nil {
		panic(err)
	}
	return &TwitchVerifier{
		Timeout:  60 * time.Second,
		provider: provider,
	}
}

type TwitchVerifier struct {
	Timeout  time.Duration
	provider *oidc.Provider
}

type TwitchVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

func (t *TwitchVerifier) ID() string {
	return "twitch"
}

func (t *TwitchVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (t *TwitchVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p TwitchVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = t.CleanToken(p.IDToken)

	if p.IDToken == "" || p.UserID == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	verifier := t.provider.Verifier(&oidc.Config{ClientID: params.ClientID})

	token, err := verifier.Verify(context.TODO(), p.IDToken)
	if err != nil {
		return false, "", err
	}

	if token.IssuedAt.Add(t.Timeout).Before(time.Now()) {
		return false, "", errors.New("timesigned is more than 60 seconds ago " + token.IssuedAt.String())
	}

	var claims map[string]interface{}
	err = token.Claims(&claims)
	if err != nil {
		return false, "", err
	}

	email, ok := claims["email"].(string)
	if !ok {
		return false, "", errors.New("email_not_found")
	}

	emailVerified, ok := claims["email_verified"].(bool)
	if !ok {
		return false, "", errors.New("email_verified_not_found")
	}

	if !emailVerified {
		return false, "", errors.New("email_not_verified")
	}

	if email != p.UserID {
		return false, "", errors.New("user_id_mismatch")
	}

	return true, p.UserID, nil
}
