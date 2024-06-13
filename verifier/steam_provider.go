package verifier

import (
	"errors"
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
	res, err := req.R().SetSuccessResult(&body).Get(provider.Endpoint + p.IDToken)
	if err != nil {
		return false, "", err
	}

	if res.IsErrorState() {
		return false, "", errors.New("steam auth api returned error")
	}

	if p.UserID != body.ID {
		return false, "", errors.New("id not equal to body.steamid " + p.UserID + " " + body.ID)
	}

	if params.ClientID != body.AZP {
		return false, "", errors.New("client id not equal to AZP " + params.ClientID + " " + body.AZP)
	}

	return true, p.UserID, nil
}
