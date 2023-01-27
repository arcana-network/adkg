package verifier

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/torusresearch/bijson"
)

type DiscordAuthResponse struct {
	Application struct {
		ID string `json:"id"`
	} `json:"application"`
	User struct {
		ID string `json:"id"`
	} `json:"user"`
	Expires string `json:"expires"`
}
type DiscordUserResponse struct {
	Verified bool   `json:"verified"`
	Email    string `json:"email"`
	ID       string `json:"id"`
}

type DiscordVerifier struct {
	Timeout time.Duration
}

type DiscordVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

func (d *DiscordVerifier) ID() string {
	return "discord"
}

func (d *DiscordVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (d *DiscordVerifier) Verify(rawPayload *bijson.RawMessage, clientID string) (bool, string, error) {
	var p DiscordVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	var body DiscordAuthResponse
	if err := getDiscordAuth(&body, p.IDToken); err != nil {
		return false, "", err
	}
	if err := verifyDiscordAuthResponse(body, p.UserID, d.Timeout, clientID); err != nil {
		return false, "", err
	}

	var user DiscordUserResponse
	if err := getDiscordEmail(&user, p.IDToken); err != nil {
		return false, "", err
	}
	if err := verifyDiscordUserResponse(user, p.UserID); err != nil {
		return false, "", fmt.Errorf("verify_discord_response: %w", err)
	}

	return true, p.UserID, nil
}

func NewDiscordProvider() *DiscordVerifier {
	return &DiscordVerifier{
		Timeout: 60 * time.Second,
	}
}

func getDiscordAuth(body *DiscordAuthResponse, idToken string) error {
	req, err := http.NewRequest("GET", "https://discordapp.com/api/oauth2/@me", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := bijson.Unmarshal(b, body); err != nil {
		return err
	}

	return nil
}
func getDiscordEmail(body *DiscordUserResponse, idToken string) error {
	req, err := http.NewRequest("GET", "https://discordapp.com/api/users/@me", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := bijson.Unmarshal(b, body); err != nil {
		return err
	}

	return nil
}

func verifyDiscordAuthResponse(body DiscordAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	timeExpires, err := time.Parse(time.RFC3339, body.Expires)
	if err != nil {
		return err
	}
	timeSigned := timeExpires.Add(-7 * 24 * time.Hour)
	if timeSigned.Add(timeout).Before(time.Now()) {
		return errors.New("timesigned is more than 60 seconds ago " + timeSigned.String())
	}

	if clientID != body.Application.ID {
		return fmt.Errorf("client ID mismatch: %s %s", clientID, body.Application.ID)
	}
	return nil
}

func verifyDiscordUserResponse(body DiscordUserResponse, verifierID string) error {
	if !body.Verified {
		return ErrorIDNotVerified
	}
	if body.ID == verifierID {
		return nil
	}
	if verifierID != body.Email {
		return fmt.Errorf("ids do not match %s %s", verifierID, body.Email)
	}
	return nil
}
