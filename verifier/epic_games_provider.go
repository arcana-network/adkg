package verifier

import (
	"context"
	"errors"
	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

type EpicGamesVerifier struct {
	Timeout  time.Duration
	provider *oidc.Provider
}

type EpicGamesVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

func NewEpicGamesVerifier() *EpicGamesVerifier {
	provider, err := oidc.NewProvider(context.TODO(), "https://api.epicgames.dev/epic/oauth/v1")
	if err != nil {
		panic(err)
	}
	return &EpicGamesVerifier{
		Timeout:  60 * time.Second,
		provider: provider,
	}
}

func (e *EpicGamesVerifier) ID() string {
	return "epic_games"
}

func (e *EpicGamesVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (e *EpicGamesVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p EpicGamesVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = e.CleanToken(p.IDToken)

	if p.IDToken == "" || p.UserID == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	verifier := e.provider.Verifier(&oidc.Config{ClientID: params.ClientID})

	token, err := verifier.Verify(context.TODO(), p.IDToken)
	if err != nil {
		return false, "", err
	}

	if token.IssuedAt.Add(e.Timeout).Before(time.Now()) {
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
