package verifier

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
)

func NewDiscordProvider() *DiscordVerifier {
	return &DiscordVerifier{
		Timeout: 60 * time.Second,
	}
}

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

func (d *DiscordVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p DiscordVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	var body DiscordAuthResponse
	res, err := req.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.IDToken)).
		SetSuccessResult(&body).
		Get("https://discordapp.com/api/oauth2/@me")
	if err != nil {
		return false, "", err
	}
	if res.IsErrorState() {
		return false, "", errors.New("discord_auth_error")
	}

	timeExpires, err := time.Parse(time.RFC3339, body.Expires)
	if err != nil {
		return false, "", err
	}
	timeSigned := timeExpires.Add(-7 * 24 * time.Hour)
	if timeSigned.Add(d.Timeout).Before(time.Now()) {
		return false, "", errors.New("token_expired")
	}

	if params.ClientID != body.Application.ID {
		return false, "", errors.New("client_id_mismatch")
	}

	var user DiscordUserResponse
	resp, err := req.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.IDToken)).
		SetSuccessResult(&user).
		Get("https://discordapp.com/api/users/@me")
	if err != nil {
		return false, "", err
	}

	if resp.IsErrorState() {
		return false, "", errors.New("discord_user_error")
	}

	if !user.Verified {
		return false, "", ErrorIDNotVerified
	}

	if p.UserID != user.Email {
		return false, "", errors.New("user_id_mismatch")
	}

	return true, p.UserID, nil
}
