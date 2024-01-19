package verifier

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"time"

	"github.com/arcana-network/dkgnode/config"
	log "github.com/sirupsen/logrus"

	"github.com/arcana-network/dkgnode/common"
	"github.com/multiformats/go-multibase"
	"github.com/smallstep/pkcs7"
	"github.com/torusresearch/bijson"
)

type GlobalKeyMessage struct {
	Verifier      string    `json:"verifier"`
	ApplicationID string    `json:"application_id"`
	ClientID      string    `json:"client_id"`
	IDToken       string    `json:"id_token"`
	AccessToken   string    `json:"access_token"`
	TokenType     string    `json:"token_type"`
	RefreshToken  string    `json:"refresh_token"`
	Expiry        time.Time `json:"expiry,omitempty"`
}

type GlobalKeyVerifier struct {
	vs               *VerifierService
	certPool         *x509.CertPool
	allowedProviders []string
}

func NewGlobalKeyVerifier(vs *VerifierService) *GlobalKeyVerifier {
	certPool := x509.NewCertPool()
	// TODO: Fetch this via API, should have an expiry also
	if !certPool.AppendCertsFromPEM([]byte(config.GlobalConfig.GlobalKeyCertPool)) {
		panic(errors.New("invalid certificate pool data"))
	}
	return &GlobalKeyVerifier{
		vs:               vs,
		certPool:         certPool,
		allowedProviders: []string{"google", "twitch"},
	}
}

func (gkv *GlobalKeyVerifier) ID() string {
	return "global_key_proxy"
}

func (gkv *GlobalKeyVerifier) CleanToken(x string) string {
	return x
}

func (gkv *GlobalKeyVerifier) Verify(rawPayload *bijson.RawMessage, _ *common.VerifierParams) (bool, string, error) {
	var p common.GenericVerifierData
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	_, decoded, err := multibase.Decode(p.Token)
	if err != nil {
		log.WithError(err).Error("GlobalProxyProvider:Multibase.decode")
		return false, "", err
	}
	msg, err := pkcs7.Parse(decoded)
	if err != nil {
		log.WithError(err).Error("GlobalProxyProvider:Pkcs7.Parse")
		return false, "", err
	}
	err = msg.VerifyWithChain(gkv.certPool)
	if err != nil {
		log.WithError(err).Error("GlobalProxyProvider:VerifyWithChain")
		return false, "", err
	}
	parsed := new(GlobalKeyMessage)
	err = json.Unmarshal(msg.Content, parsed)
	if err != nil {
		log.WithError(err).Error("GlobalProxyProvider:unmarshal")
		return false, "", err
	}

	if p.AppID != parsed.ApplicationID {
		return false, "", errors.New("application ID did not match")
	}
	serialized, err := bijson.Marshal(common.GenericVerifierData{
		UserID:   p.UserID,
		Token:    parsed.IDToken,
		Provider: parsed.Verifier,
		AppID:    parsed.ApplicationID,
	})

	return gkv.vs.providerMap.Verify((*bijson.RawMessage)(&serialized), serviceMapper)
}
