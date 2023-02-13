package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/secret"
	"github.com/arcana-network/dkgnode/secret/vault"
)

var GlobalConfig *Config

type Config struct {
	TMP2PListenAddress string `json:"tmp2plistenaddress"`
	P2PListenAddress   string `json:"p2plistenaddress"`
	PrivateKey         []byte
	TMPrivateKey       []byte
	SecretConfigPath   string `json:"secretConfigPath"`
	BasePath           string `json:"dataDirectory"`
	IPAddress          string `json:"ipAddress"`
	EthConnection      string `json:"blockchainRPCURL"`
	ContractAddress    string `json:"dkgContractAddress"`
	HttpServerPort     string `json:"port"`
	Domain             string `json:"domain"`
	GatewayURL         string `json:"gatewayUrl"`
	PasswordlessUrl    string `json:"passwordlessUrl"`
	OAuthUrl           string `json:"oauthUrl"`
}

func (c *Config) VerifyRequired() error {
	if c.SecretConfigPath == "" {
		return errors.New("required secretConfigPath missing")
	}
	if c.IPAddress == "" {
		return errors.New("required ipAddress missing")
	}
	return nil
}

func ConfigFromFile(configPath string) (*Config, error) {
	config, err := ReadConfigJson(configPath)
	if err != nil {
		return nil, err
	}

	if config.IPAddress != "" {
		UseIPAdressInListenAddress(config)
	}

	return config, nil
}

func UseIPAdressInListenAddress(config *Config) {
	config.TMP2PListenAddress = fmt.Sprintf("tcp://%s:26656", config.IPAddress)
	config.P2PListenAddress = fmt.Sprintf("/ip4/%s/tcp/1080", config.IPAddress)
}

func ReadConfigJson(configPath string) (*Config, error) {
	config := GetDefaultConfig()
	f, err := os.OpenFile(configPath, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}
	return config, nil
}

func GetDefaultConfig() *Config {
	config := &Config{
		TMP2PListenAddress: "tcp://0.0.0.0:26656",
		P2PListenAddress:   "/ip4/0.0.0.0/tcp/1080",
		BasePath:           "/tmp/keygen-data",
		EthConnection:      DefaultBlockchainRPCURL,
		ContractAddress:    DefaultContractAddress,
		GatewayURL:         DefaultGatewayURL,
	}
	return config
}

func GetPrivateKeys(configPath string) (key secret.Keys, err error) {
	c, err := secret.ReadConfig(configPath)
	if err != nil {
		return
	}

	manager, err := vault.NewVaultManager(c)
	if err != nil {
		return
	}

	err = manager.Setup()
	if err != nil {
		return
	}

	nodePrivKey, err := manager.GetSecret(secret.NodeKey)
	if err != nil {
		return
	}
	key.NodePrivateKey = nodePrivKey

	tmPrivKey, err := manager.GetSecret(secret.TendermintKey)
	if err != nil {
		return
	}
	key.TmPrivateKey = tmPrivKey
	return
}
