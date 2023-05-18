package verifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	log "github.com/sirupsen/logrus"
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

type SteamProviderResponse struct {
	ID string `json:"id"`
}

func (g *SteamProvider) ID() string {
	return "steam"
}

func (g *SteamProvider) CleanToken(token string) string {
	return strings.Trim(token, " ")
}

func (g *SteamProvider) Verify(rawPayload *bijson.RawMessage, params *common.VerifierParams) (bool, string, error) {
	var p SteamProviderParams
	if err := bijson.Unmarshal(*rawPayload, &p); err != nil {
		return false, "", err
	}

	p.IDToken = g.CleanToken(p.IDToken)
	if p.UserID == "" || p.IDToken == "" {
		return false, "", errors.New("invalid payload parameters")
	}

	url := g.Endpoint + p.IDToken
	var body SteamProviderResponse
	err := getSteamAuth(url, &body)

	if err != nil {
		return false, "", err
	}

	err = verifySteamAuthResponse(body, p.UserID, g.Timeout, params.ClientID)
	if err != nil {
		return false, "", fmt.Errorf("verify_google_response: %w", err)
	}

	return true, p.UserID, nil
}

func getSteamAuth(url string, body *SteamProviderResponse) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"StatusCode": resp.StatusCode,
		"HTTPStatus": resp.Status,
	}).Debugf("GoogleVerifier")
	if resp.StatusCode >= 400 {
		return fmt.Errorf("error from google auth. code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, body)
	if err != nil {
		return err
	}
	return nil
}

func verifySteamAuthResponse(body SteamProviderResponse, verifierID string, timeout time.Duration, clientID string) error {
	if strings.Compare(verifierID, body.ID) != 0 {
		return errors.New("email not equal to body.steamid " + verifierID + " " + body.ID)
	}
	return nil
}
