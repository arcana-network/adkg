package verifier

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/imroc/req/v3"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type PasswordlessVerifier struct {
	Version  string
	Endpoint string
	Timeout  time.Duration
}

func NewPasswordlessProvider() *PasswordlessVerifier {
	verifyUrl, err := url.Parse(config.GlobalConfig.PasswordlessUrl)
	if err != nil {
		panic(err)
	}
	verifyUrl.Path = "api/token/verify"
	verifyUrl.RawQuery = "id_token="
	return &PasswordlessVerifier{
		Version:  "1.0",
		Endpoint: verifyUrl.String(),
		Timeout:  600 * time.Second,
	}
}

type PasswordlessVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

type PasswordlessAuthResponse struct {
	Email string   `json:"email"`
	Iat   int      `json:"iat"`
	Exp   int      `json:"exp"`
	Azp   string   `json:"azp"`
	Iss   string   `json:"iss"`
	Aud   []string `json:"aud"`
	Sub   string   `json:"sub"`
}

func (g *PasswordlessVerifier) ID() string {
	return "passwordless"
}

func (g *PasswordlessVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (g *PasswordlessVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p PasswordlessVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = g.CleanToken(p.IDToken)
	if p.UserID == "" || p.IDToken == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	var body PasswordlessAuthResponse

	if _, err := req.R().
		SetSuccessResult(&body).
		Get(g.Endpoint + p.IDToken); err != nil {
		return false, "", err
	}

	if err := verifyPasswordlessResponse(body, p.UserID, g.Timeout, params.ClientID); err != nil {
		log.WithError(err).Error("PasswordlessVerifier:Verify")
		return false, "", fmt.Errorf("error: %w", err)
	}

	return true, p.UserID, nil
}

func verifyPasswordlessResponse(body PasswordlessAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	timeSigned := time.Unix(int64(body.Iat), 0)
	if timeSigned.Add(timeout).Before(time.Now()) {
		return errors.New("timesigned is more than 60 seconds ago " + timeSigned.String())
	}
	if strings.Compare(verifierID, body.Email) != 0 {
		return errors.New("email not equal to body.email " + verifierID + " " + body.Email)
	}
	if strings.Compare(clientID, body.Azp) != 0 {
		return fmt.Errorf("clientID Mismatch: Expected=%s, Got=%s", clientID, body.Azp)
	}
	return nil
}
