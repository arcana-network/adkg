package verifier

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
)

func NewGithubProvider() *GithubVerifier {
	return &GithubVerifier{
		Timeout: 60 * time.Second,
	}
}

type GithubAuthResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GithubVerifier struct {
	Timeout time.Duration
}

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
		return false, "", errors.New("github_auth_error")
	}

	idStr := strconv.FormatInt(int64(body.ID), 10)

	if idStr != p.UserID && body.Email != p.UserID {
		return false, "", fmt.Errorf("user_id_mismatch")
	}

	return true, p.UserID, nil
}
