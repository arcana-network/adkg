package verifier

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/imroc/req/v3"
	"github.com/torusresearch/bijson"
)

type SteamProvider struct {
	Version  string
	Endpoint string
	Timeout  time.Duration
}

func NewSteamProvider() *SteamProvider {
	endpoint, _ := url.Parse(config.GlobalConfig.OAuthUrl)
	endpoint.Path = "/api/steam/verify"
	endpoint.RawQuery = "token="

	return &SteamProvider{
		Version:  "1.0",
		Endpoint: endpoint.String(),
		Timeout:  600 * time.Second,
	}
}

type SteamProviderParams struct {
	IDToken string `json:"id_token"`
	UserID  string `json:"user_id"`
}

type SteamAuthResponse struct {
	ID  string `json:"id"`
	AZP string `json:"azp"`
}

func (p *SteamProvider) ID() string {
	return "steam"
}

func (p *SteamProvider) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (provider *SteamProvider) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p SteamProviderParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = provider.CleanToken(p.IDToken)
	if p.UserID == "" || p.IDToken == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	var body SteamAuthResponse
	if _, err := req.R().SetSuccessResult(&body).Get(provider.Endpoint + p.IDToken); err != nil {
		return false, "", err
	}

	if err := verifySteamResponse(body, p.UserID, provider.Timeout, params.ClientID); err != nil {
		return false, "", fmt.Errorf("verify_steam_response: %w", err)
	}

	return true, p.UserID, nil
}

func verifySteamResponse(body SteamAuthResponse, verifierID string, timeout time.Duration, clientID string) error {
	if strings.Compare(verifierID, body.ID) != 0 {
		return errors.New("id not equal to body.steamid " + verifierID + " " + body.ID)
	}

	if strings.Compare(clientID, body.AZP) != 0 {
		return errors.New("client id not equal to AZP " + clientID + " " + body.AZP)
	}

	return nil
}
