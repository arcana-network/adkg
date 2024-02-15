package verifier

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
)

type GoogleVerifier struct {
	Version  string
	Endpoint string
	Timeout  time.Duration
}

func NewGoogleProvider() *GoogleVerifier {
	return &GoogleVerifier{
		Version:  "1.0",
		Endpoint: "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=",
		Timeout:  600 * time.Second,
	}
}

type GoogleVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

type GoogleAuthResponse struct {
	Azp           string `json:"azp"`
	Email         string `json:"email"`
	Iss           string `json:"iss"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	EmailVerified string `json:"email_verified"`
	AtHash        string `json:"at_hash"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	Locale        string `json:"locale"`
	Iat           string `json:"iat"`
	Exp           string `json:"exp"`
	Jti           string `json:"jti"`
	Alg           string `json:"alg"`
	Kid           string `json:"kid"`
	Typ           string `json:"typ"`
}

func (g *GoogleVerifier) ID() string {
	return "google"
}

func (g *GoogleVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (g *GoogleVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p GoogleVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = g.CleanToken(p.IDToken)

	if p.UserID == "" || p.IDToken == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	var body GoogleAuthResponse
	res, err := req.R().
		SetSuccessResult(&body).
		Get(g.Endpoint + p.IDToken)
	if err != nil {
		return false, "", err
	}

	if res.IsErrorState() {
		return false, "", errors.New("google auth api returned error")
	}

	if err := verifyGoogleResponse(body, p.UserID, g.Timeout, params.ClientID); err != nil {
		return false, "", fmt.Errorf("verify_google_response: %w", err)
	}

	return true, p.UserID, nil
}

func verifyGoogleResponse(body GoogleAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	timeSignedInt, ok := new(big.Int).SetString(body.Iat, 10)
	if !ok {
		return errors.New("Could not get timesignedint from " + body.Iat)
	}
	if body.EmailVerified != "true" {
		return ErrorIDNotVerified
	}
	timeSigned := time.Unix(timeSignedInt.Int64(), 0)

	if timeSigned.Add(timeout).Before(time.Now()) {
		return errors.New("timesigned is more than 60 seconds ago " + timeSigned.String())
	}
	if verifierID != body.Email {
		return errors.New("email not equal to body.email " + verifierID + " " + body.Email)
	}
	if clientID != body.Azp {
		return fmt.Errorf("clientID mismatch: Expected:%s Got:%s", clientID, body.Azp)
	}
	return nil
}
