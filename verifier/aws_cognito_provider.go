package verifier

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/torusresearch/bijson"
)

type AWSCognitoVerifier struct {
	// in seconds
	Timeout int64
}

func NewAWSCognitoVerifier() *AWSCognitoVerifier {
	return &AWSCognitoVerifier{
		Timeout: 600,
	}
}

type AWSCognitoVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

type AWSCognitoAuthResponse struct {
	/*
		AZP           string `json:"azp"`
		ISS           string `json:"iss"`
		IAT           int64  `json:"iat"`
		AUD           string `json:"aud"`
		SUB           string `json:"sub"`
	*/
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (a *AWSCognitoVerifier) ID() string {
	return "aws_cognito"
}

func (a *AWSCognitoVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

// AWS Cognito domains are user controlled, therefore errors are not worth saving.
func (a *AWSCognitoVerifier) callAndVerify(p AWSCognitoVerifierParams, domain, clientID string) bool {
	// Domain is uncontrolled unfortunately, it can be practically anything
	req, err := http.NewRequest("GET", "https://"+domain+"/oauth2/userinfo", nil)
	if err != nil {
		return false
	}
	req.Header.Add("Authorization", "Bearer "+p.IDToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode >= 400 {
		return false
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var authResp AWSCognitoAuthResponse
	if err := bijson.Unmarshal(b, &authResp); err != nil {
		return false
	}

	/*
		if authResp.IAT < time.Now().Unix()-a.Timeout {
			return false
		}

		if authResp.AZP != clientID {
			return false
		}
	*/

	if authResp.Email != p.UserID || authResp.EmailVerified != "true" {
		return false
	}

	return true
}

func (a *AWSCognitoVerifier) Verify(rawPayload *bijson.RawMessage, joinedClientID string) (bool, string, error) {
	var domain, clientID string
	{
		split := strings.Split(joinedClientID, "|")
		if len(split) != 2 {
			return false, "", errors.New("invalid client ID")
		}
		domain = split[0]
		clientID = split[1]
	}
	var p AWSCognitoVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}
	ok := a.callAndVerify(p, domain, clientID)
	if !ok {
		return false, p.UserID, nil
	}

	return true, "", nil
}
