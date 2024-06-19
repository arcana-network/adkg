package verifier

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/eventbus"
	"github.com/torusresearch/bijson"
)

var ErrorIDNotVerified = errors.New("user_not_verified")

type VerifierService struct {
	providerMap *ProviderMap
	bus         eventbus.Bus
}

type VerifyMessage struct {
	Token    string `json:"id_token"`
	Provider string `json:"provider"`
	AppID    string `json:"app_id"`
}

type Provider interface {
	ID() string
	CleanToken(string) string
	Verify(*bijson.RawMessage, *common.VerifierParams) (verified bool, verifierID string, err error)
}

type ProviderMap struct {
	Providers map[string]Provider
}

var serviceMapper *common.MessageBroker

func (tgv *ProviderMap) Verify(rawMessage *bijson.RawMessage, serviceMapper *common.MessageBroker) (bool, string, error) {
	var msg common.GenericVerifierData
	if err := bijson.Unmarshal(*rawMessage, &msg); err != nil {
		return false, "", err
	}
	v, err := tgv.Lookup(msg.Provider)
	if err != nil {
		return false, "", err
	}

	if msg.Provider == "global_key_proxy" {
		return v.Verify(rawMessage, nil)
	}

	cleanedToken := v.CleanToken(msg.Token)
	if cleanedToken != msg.Token {
		return false, "", errors.New("cleaned token is different from original token")
	}
	params, err := getVerifierParams(serviceMapper, msg.AppID, msg.Provider)
	if err != nil || params == nil || params.ClientID == "" {
		return false, "", errors.New("invalid app address")
	}

	return v.Verify(rawMessage, params)
}

func getVerifierParams(serviceMapper *common.MessageBroker, appID, verifier string) (*common.VerifierParams, error) {
	params, err := serviceMapper.ChainMethods().GetClientIDViaVerifier(appID, verifier)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func (tgv *ProviderMap) Lookup(provider string) (Provider, error) {
	if tgv.Providers == nil {
		return nil, errors.New("providers mapping not initialized")
	}
	if tgv.Providers[provider] == nil {
		if strings.HasPrefix(provider, "csp-") {
			return tgv.Providers["custom"], nil
		}
		return nil, errors.New("provider:" + provider + " not found")
	}
	return tgv.Providers[provider], nil
}

func NewProviderMap(providers []Provider) *ProviderMap {
	providerMap := &ProviderMap{
		Providers: make(map[string]Provider),
	}
	for _, provider := range providers {
		providerMap.Providers[provider.ID()] = provider
	}
	return providerMap
}

func New(bus eventbus.Bus) *VerifierService {
	verifierService := VerifierService{
		bus: bus,
	}
	serviceMapper = common.NewServiceBroker(bus, common.VERIFIER_SERVICE_NAME)
	return &verifierService
}

func (*VerifierService) ID() string {
	return common.VERIFIER_SERVICE_NAME
}

func (v *VerifierService) Start() error {
	firebaseProvider, err := NewFirebaseProvider()
	if err != nil {
		return err
	}

	providers := []Provider{
		NewGoogleProvider(),
		NewDiscordProvider(),
		NewTwitchProvider(),
		NewGithubProvider(),
		NewTwitterProvider(),
		NewPasswordlessProvider(),
		NewAWSCognitoVerifier(),
		NewSteamProvider(),
		firebaseProvider,
		NewGlobalKeyVerifier(v),
		NewCustomProvider(),
		// NewXProvider(),
		NewAppleIDProvider(),
	}
	v.providerMap = NewProviderMap(providers)
	return nil

}
func (v *VerifierService) Stop() error {
	return nil
}
func (v *VerifierService) IsRunning() bool {
	return true
}
func (v *VerifierService) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "verify":
		var msg bijson.RawMessage
		_ = common.CastOrUnmarshal(args[0], &msg)
		rs := new(struct {
			Valid  bool
			UserID string
		})
		serviceMapper := common.NewServiceBroker(v.bus, "verifier")
		valid, userID, err := v.providerMap.Verify(&msg, serviceMapper)
		rs.Valid = valid
		rs.UserID = userID
		return *rs, err
	case "clean_token":
		var provider, token string
		_ = common.CastOrUnmarshal(args[0], &provider)
		_ = common.CastOrUnmarshal(args[1], &token)

		verifier, err := v.providerMap.Lookup(provider)
		if err != nil {
			return nil, err
		}
		return verifier.CleanToken(token), nil
	}
	return nil, fmt.Errorf("verifier service method %v not found", method)
}
