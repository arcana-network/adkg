package verifier

import (
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/imroc/req/v3"
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
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func (a *AWSCognitoVerifier) ID() string {
	return "aws"
}

func (a *AWSCognitoVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

// AWS Cognito domains are user controlled, therefore errors are not worth saving.
func (a *AWSCognitoVerifier) callAndVerify(p AWSCognitoVerifierParams, params *common.VerifierParams) bool {
	// Domain is uncontrolled unfortunately, it can be practically anything
	url := "https://" + params.Domain + "/oauth2/userInfo"
	// req.DevMode()

	var authResp AWSCognitoAuthResponse
	_, err := req.R().
		SetHeader("Authorization", "Bearer "+p.IDToken).
		SetSuccessResult(&authResp).
		Get(url)
	if err != nil {
		// panic("№1")
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

	if authResp.Sub != p.UserID || authResp.EmailVerified != "true" {
		// panic("№2")
		return false
	}

	return true
}

func (a *AWSCognitoVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p AWSCognitoVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}
	ok := a.callAndVerify(p, params)
	if !ok {
		return false, "", nil
	}

	return true, p.UserID, nil
}
