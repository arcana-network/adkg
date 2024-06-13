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

func NewXProvider() *XVerifier {
	return &XVerifier{
		Timeout: 60 * time.Second,
		URL:     "https://api.twitter.com/2/users/me",
	}
}

type XVerifier struct {
	Timeout time.Duration
	URL     string
}

type XInputParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

type XAPIResponse struct {
	ID string `json:"id"`
}

func (t *XVerifier) ID() string {
	return "x"
}

func (t *XVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (t *XVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p XInputParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = t.CleanToken(p.IDToken)

	if p.IDToken == "" || p.UserID == "" {
		return false, "", errors.New("invalid input parameters")
	}

	var body XAPIResponse

	res, err := req.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.IDToken)).
		SetSuccessResult(body).
		Get(t.URL)
	if err != nil {
		return false, "", err
	}

	if res.IsErrorState() {
		return false, "", errors.New("twitter api returned error")
	}

	if body.ID != p.UserID {
		return false, "", errors.New("user id mismatch")
	}

	return true, p.UserID, nil
}
