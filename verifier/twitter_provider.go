package verifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"

	"github.com/torusresearch/bijson"

	"github.com/arcana-network/dkgnode/keygen"
)

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

func (t *TwitterVerifier) Verify(rawPayload *bijson.RawMessage, clientID string) (bool, string, error) {
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

	if err := verifyTwitterAuthResponse(body, p.UserID, t.Timeout, clientID); err != nil {
		return false, "", fmt.Errorf("verify_twitter_response: %w", err)
	}

	return true, p.UserID, nil
}

func NewTwitterProvider() *TwitterVerifier {
	return &TwitterVerifier{
		Timeout: 60 * time.Second,
		// SignatureUrl: "http://host.docker.internal:9000/oauth/twitter/signature",
		SignatureUrl: "https://api-auth.arcana.network/oauth/twitter/signature",
		UserInfoUrl:  "https://api.twitter.com/1.1/account/verify_credentials.json",
	}
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
	sig := keygen.ECDSASign(getSignatureMessage.String(), &privKey)
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
		log.WithField("Error", err).Info("MarshalBody: TwitterAuthentication")
		return nil, errors.New("error marshalling body for signature")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.WithField("Error", err).Info("GetSignature: TwitterAuthentication")
		return nil, errors.New("error creating req for signature")
	}
	var body TwitterSignatureResponse

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithField("Error", err).Info("GetSignature: TwitterAuthentication")
		return nil, errors.New("error getting signature from secret keeper")
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithField("Error", err).Info("ReadBody: TwitterAuthentication")
		return nil, errors.New("error reading body from secret keeper")
	}
	if err := bijson.Unmarshal(b, &body); err != nil {
		return nil, err
	}
	return &body, nil
}

func getTwitterAuth(body *TwitterAuthResponse, v *TwitterVerifier, idToken string) error {
	log.WithField("idToken", idToken).Info("IDToken:TwitterVerifier")
	s := strings.Split(idToken, ":")
	if len(s) != 3 {
		return errors.New("unexpected id token")
	}
	sigBody := SignatureBody{
		OauthToken:       s[0],
		OauthTokenSecret: s[1],
		AppID:            s[2],
	}
	log.WithField("sigBody: ", sigBody).Info("getTwitterAuth:TwitterVerifier")

	sig, err := getSignedRequest(v.SignatureUrl, sigBody)
	if err != nil {
		return err
	}

	log.WithField("sig: ", sig).Info("getSignedRequest:getTwitterAuth:TwitterVerifier")

	requestUrl, _ := url.Parse(v.UserInfoUrl)
	requestUrl.RawQuery = url.Values(sig.Params).Encode()

	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		log.WithField("err", err).Info("CreateRequest:TwitterVerifier")
		return err
	}

	req.Header.Add("Authorization", sig.Header)

	bytesRes, err := request(req)
	if err != nil {
		log.WithField("err", err).Info("VerifyCredentials:TwitterVerifier")
		return err
	}

	if err := json.Unmarshal(bytesRes, &body); err != nil {
		log.WithField("err", err).Info("ParsingTwitterResponse:TwitterVerifier")
		return err
	}
	return nil
}

func request(req *http.Request) (bodyBytes []byte, err error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.WithField("err", err).Info("DoRequest:TwitterVerifier")
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("server returned status error")
		return
	}
	defer resp.Body.Close()
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.WithField("err", err).Info("ReadRequestResponse:TwitterVerifier")
		return
	}
	return bodyBytes, err
}
func verifyTwitterAuthResponse(body TwitterAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	log.WithField("body", body).Info("Twitter verifier")
	if body.ID != verifierID && body.Email != verifierID {
		return fmt.Errorf("UserIDs do not match: %s %s", body.ID, verifierID)
	}
	return nil
}
