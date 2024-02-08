package verifier

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/imroc/req/v3"

	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/config"
)

func NewTwitterProvider() *TwitterVerifier {
	signatureUrl, err := url.Parse(config.GlobalConfig.OAuthUrl)
	if err != nil {
		panic(err)
	}
	signatureUrl.Path = "/oauth/twitter/signature"

	return &TwitterVerifier{
		Timeout:      60 * time.Second,
		SignatureUrl: signatureUrl.String(),
		UserInfoUrl:  "https://api.twitter.com/1.1/account/verify_credentials.json",
	}
}

type TwitterAuthResponse struct {
	ID           string `json:"id_str"`
	ProfileImage string `json:"profile_image_url_https"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}
type TwitterSignatureResponse struct {
	Params map[string][]string `json:"params"`
	Url    string              `json:"url"`
	Header string              `json:"header"`
}

type TwitterVerifier struct {
	Timeout      time.Duration
	SignatureUrl string
	UserInfoUrl  string
}

type TwitterVerifierParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

func (t *TwitterVerifier) ID() string {
	return "twitter"
}

func (t *TwitterVerifier) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (t *TwitterVerifier) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p TwitterVerifierParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = t.CleanToken(p.IDToken)

	if p.IDToken == "" || p.UserID == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	var body TwitterAuthResponse

	if err := getTwitterAuth(&body, t, p.IDToken); err != nil {
		return false, "", err
	}

	if body.ID != p.UserID && body.Email != p.UserID {
		return false, "", errors.New("user_id_mismatch")
	}

	return true, p.UserID, nil
}

type SignatureBody struct {
	OauthToken       string `json:"oauth_token"`
	OauthTokenSecret string `json:"oauth_token_secret"`
	AppID            string `json:"app_id"`
}

type GetSignatureParams struct {
	PubKeyX             string              `json:"pubkeyx"`
	PubKeyY             string              `json:"pubkeyy"`
	GetSignatureMessage GetSignatureMessage `json:"get_signature_message"`
	Signature           []byte              `json:"signature"`
}
type GetSignatureMessage struct {
	Timestamp        string      `json:"timestamp"`
	OauthToken       string      `json:"oauth_token"`
	OauthTokenSecret string      `json:"oauth_token_secret"`
	AppID            string      `json:"app_id"`
	NodeAddress      eth.Address `json:"node_address"`
}

func (c *GetSignatureMessage) String() string {
	Delimiter1 := "\x1c"
	params := []string{
		c.Timestamp,
		c.OauthToken,
		c.OauthTokenSecret,
		c.AppID,
		c.NodeAddress.String(),
	}
	return strings.Join(params, Delimiter1)
}

func getSignatureParams(token, secret, appID string) ([]byte, error) {
	pubKey := serviceMapper.ChainMethods().GetSelfPublicKey()
	privKey := serviceMapper.ChainMethods().GetSelfPrivateKey()
	getSignatureMessage := GetSignatureMessage{
		OauthToken:       token,
		OauthTokenSecret: secret,
		AppID:            appID,
		Timestamp:        strconv.FormatInt(time.Now().Unix(), 10),
		NodeAddress:      serviceMapper.ChainMethods().GetSelfAddress(),
	}
	sig := common.ECDSASign(getSignatureMessage.String(), &privKey)
	getSigParams := GetSignatureParams{
		PubKeyX:             pubKey.X.Text(16),
		PubKeyY:             pubKey.Y.Text(16),
		GetSignatureMessage: getSignatureMessage,
		Signature:           sig,
	}
	body, err := bijson.Marshal(getSigParams)
	return body, err
}

func getSignedRequest(url string, sigBody SignatureBody) (*TwitterSignatureResponse, error) {
	bodyBytes, err := getSignatureParams(sigBody.OauthToken, sigBody.OauthTokenSecret, sigBody.AppID)
	if err != nil {
		return nil, errors.New("error marshalling body for signature")
	}

	var body TwitterSignatureResponse
	res, err := req.R().
		SetBodyBytes(bodyBytes).
		SetSuccessResult(&body).
		Post(url)
	if err != nil {
		return nil, errors.New("twitch_sig_error")
	}

	if res.IsErrorState() {
		return nil, errors.New("twitch_sig_error")
	}
	return &body, nil
}

func getTwitterAuth(body *TwitterAuthResponse, v *TwitterVerifier, idToken string) error {
	s := strings.Split(idToken, ":")
	if len(s) != 3 {
		return errors.New("unexpected id token")
	}
	sigBody := SignatureBody{
		OauthToken:       s[0],
		OauthTokenSecret: s[1],
		AppID:            s[2],
	}

	sig, err := getSignedRequest(v.SignatureUrl, sigBody)
	if err != nil {
		return err
	}

	u, _ := url.Parse(v.UserInfoUrl)
	u.RawQuery = url.Values(sig.Params).Encode()

	res, err := req.R().
		SetHeader("Authorization", sig.Header).
		SetSuccessResult(body).
		Get(u.String())
	if err != nil {
		return err
	}

	if res.IsErrorState() {
		return errors.New("twitter_auth_error")
	}

	return nil
}
