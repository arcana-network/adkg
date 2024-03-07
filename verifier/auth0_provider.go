package verifier

import (
	"github.com/arcana-network/dkgnode/common"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
	"strings"
)

type Auth0Verifier struct {
	Timeout int64
}

func NewAuth0Verifier() *Auth0Verifier {
	return &Auth0Verifier{Timeout: 120}
}

func (a *Auth0Verifier) ID() string {
	return "auth0"
}

func (a *Auth0Verifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

// AWS Cognito domains are user controlled, therefore errors are not worth saving.
func (a *Auth0Verifier) callAndVerify(p AWSCognitoVerifierParams, params *common.VerifierParams) bool {
	// Domain is uncontrolled unfortunately, it can be practically anything
	url := "https://" + params.Domain + "/userinfo"
	// req.DevMode()

	// Is similar so we can just re-use this
	var authResp AWSCognitoAuthResponse
	_, err := req.R().SetHeader("Authorization", "Bearer "+p.IDToken).SetSuccessResult(&authResp).Get(url)
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

	if authResp.Sub != p.UserID {
		// panic("№2")
		return false
	}

	return true
}

func (a *Auth0Verifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
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
