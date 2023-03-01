package verifier

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/torusresearch/bijson"
)

type TwitchAuthResponse struct {
	AUD      string `json:"aud"`
	EXP      int    `json:"exp"`
	IAT      int    `json:"iat"`
	ISS      string `json:"iss"`
	SUB      string `json:"sub"`
	AZP      string `json:"azp"`
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
}

type TwitchVerifier struct {
	Timeout time.Duration
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

	var body TwitchAuthResponse

	if err := getTwitchAuth(&body, p.IDToken); err != nil {
		return false, "", err
	}

	if err := verifyTwitchAuthResponse(body, p.UserID, t.Timeout, params.ClientID); err != nil {
		return false, "", fmt.Errorf("verify_twitch_response: %w", err)
	}

	return true, p.UserID, nil
}

func NewTwitchProvider() *TwitchVerifier {
	return &TwitchVerifier{
		Timeout: 60 * time.Second,
	}
}

func getTwitchAuth(body *TwitchAuthResponse, idToken string) error {
	req, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/userinfo", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("error from google auth. code %d", resp.StatusCode)
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

func verifyTwitchAuthResponse(body TwitchAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	timeSigned := time.Unix(int64(body.IAT), 0)

	if !body.Verified {
		return ErrorIDNotVerified
	}

	if timeSigned.Add(timeout).Before(time.Now()) {
		return errors.New("timesigned is more than 60 seconds ago " + timeSigned.String())
	}
	if body.Email != verifierID {
		return fmt.Errorf("UserIDs do not match %s %s", body.Email, verifierID)
	}
	if body.AUD != clientID {
		return fmt.Errorf("ClientIDMismatch: %s %s", clientID, body.AUD)
	}
	return nil
}
