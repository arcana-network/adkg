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

func NewTwitchProvider() *TwitchVerifier {
	return &TwitchVerifier{
		Timeout: 60 * time.Second,
	}
}

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

	res, err := req.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.IDToken)).
		SetSuccessResult(&body).
		Get("https://id.twitch.tv/oauth2/userinfo")
	if err != nil {
		return false, "", err
	}

	if res.IsErrorState() {
		return false, "", errors.New("twitch_auth_error")
	}

	timeSigned := time.Unix(int64(body.IAT), 0)

	if !body.Verified {
		return false, "", ErrorIDNotVerified
	}

	if timeSigned.Add(t.Timeout).Before(time.Now()) {
		return false, "", errors.New("timesigned is more than 60 seconds ago " + timeSigned.String())
	}
	if body.Email != p.UserID {
		return false, "", errors.New("user_id_mismatch")
	}
	if body.AUD != params.ClientID {
		return false, "", fmt.Errorf("client_id_mismatch: %s %s", params.ClientID, body.AUD)
	}

	return true, p.UserID, nil
}
