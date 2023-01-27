package verifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
)

type PasswordlessVerifier struct {
	Version  string
	Endpoint string
	Timeout  time.Duration
}

func NewPasswordlessProvider() *PasswordlessVerifier {
	return &PasswordlessVerifier{
		Version: "1.0",
		// Endpoint: "http://host.docker.internal:3000/api/token/verify?id_token=",
		Endpoint: "https://passwordless.dev.arcana.network/api/token/verify?id_token=",
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

func (g *PasswordlessVerifier) Verify(rawPayload *bijson.RawMessage, clientID string) (bool, string, error) {
	var p PasswordlessVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	log.WithField("clientID", clientID).Info("VerifyRequestIdentity-Passwordless")

	p.IDToken = g.CleanToken(p.IDToken)
	if p.UserID == "" || p.IDToken == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	log.WithFields(log.Fields{
		"params":   p,
		"verifier": g,
	}).Info("Passwordless")

	url := g.Endpoint + p.IDToken
	log.WithField("url", url).Info("TokenInfo-Passwordless")

	var body PasswordlessAuthResponse
	log.WithField("body", body).Info("EmptyBody-Passwordless")

	err := getPasswordlessAuth(url, &body)
	log.WithField("err", err).Info("GetAuth-Passwordless")

	if err != nil {
		return false, "", err
	}

	err = verifyPasswordlessAuthResponse(body, p.UserID, g.Timeout, clientID)
	log.WithField("err", err).Debug("VerifyAuth-Passwordless")

	if err != nil {
		return false, "", fmt.Errorf("verify_pwdless_response: %w", err)
	}

	return true, p.UserID, nil
}

func getPasswordlessAuth(url string, body *PasswordlessAuthResponse) error {
	log.WithField("url", url).Info("getPasswordlessAuth-PasswordlessVerifier")
	log.WithField("body", body).Info("getPasswordlessAuth-PasswordlessVerifier")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	log.WithField("Httpstatus code", resp.StatusCode).Info("PasswordlessVerifier")
	log.WithField("Httpstatus", resp.Status).Info("PasswordlessVerifier")
	if resp.StatusCode >= 400 {
		return fmt.Errorf("error from passwordless auth. code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, body)
	if err != nil {
		return err
	}
	return nil
}

func verifyPasswordlessAuthResponse(body PasswordlessAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	timeSigned := time.Unix(int64(body.Iat), 0)
	if timeSigned.Add(timeout).Before(time.Now()) {
		return errors.New("timesigned is more than 60 seconds ago " + timeSigned.String())
	}
	if strings.Compare(verifierID, body.Email) != 0 {
		return errors.New("email not equal to body.email " + verifierID + " " + body.Email)
	}
	if strings.Compare(clientID, body.Azp) != 0 {
		return fmt.Errorf("ClientIDMismatch: %s %s", clientID, body.Azp)
	}
	log.WithField("body", body).Info("verifyPasswordlessAuthResponse")
	return nil
}
