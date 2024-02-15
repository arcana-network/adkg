package verifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
)

func NewGithubProvider() *GithubVerifier {
	return &GithubVerifier{}
}

type GithubAuthResponse struct {
	Email string `json:"email"`
}

type GithubVerifier struct{}

type GithubVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

func (t *GithubVerifier) ID() string {
	return "github"
}

func (t *GithubVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (t *GithubVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p GithubVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = t.CleanToken(p.IDToken)

	if p.IDToken == "" || p.UserID == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	var body GithubAuthResponse

	res, err := req.R().
		SetHeader("Authorization", fmt.Sprintf("token %s", p.IDToken)).
		SetSuccessResult(&body).
		Get("https://api.github.com/user")
	if err != nil {
		return false, "", err
	}

	if res.IsErrorState() {
		return false, "", errors.New("github auth api returned error")
	}

	if body.Email != p.UserID {
		return false, "", fmt.Errorf("user id mismatch: only email is allowed")
	}

	return true, p.UserID, nil
}
