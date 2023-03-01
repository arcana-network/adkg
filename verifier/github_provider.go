package verifier

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

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

	if err := getGithubAuth(&body, p.IDToken); err != nil {
		return false, "", err
	}

	if err := verifyGithubAuthResponse(body, p.UserID, t.Timeout, params.ClientID); err != nil {
		return false, "", fmt.Errorf("verify_github_response: %w", err)
	}

	return true, p.UserID, nil
}

func NewGithubProvider() *GithubVerifier {
	return &GithubVerifier{
		Timeout: 60 * time.Second,
	}
}

func getGithubAuth(body *GithubAuthResponse, idToken string) error {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", idToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("error from github auth. code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := bijson.Unmarshal(b, &body); err != nil {
		return err
	}
	return nil
}

func verifyGithubAuthResponse(body GithubAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	log.WithField("body", body).Debug("GithubVerifier")
	idStr := strconv.FormatInt(int64(body.ID), 10)
	if idStr != verifierID && body.Email != verifierID {
		log.WithFields(log.Fields{
			"idStr":      idStr,
			"verifierID": verifierID,
			"email":      body.Email,
		}).Error("GithubVerify:IdMismatch")
		return fmt.Errorf("User ID did not match the one specified.")
	}
	return nil
}
