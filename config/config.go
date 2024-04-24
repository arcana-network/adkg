package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/secret"
	"github.com/arcana-network/dkgnode/secret/vault"
	log "github.com/sirupsen/logrus"
)

var GlobalConfig *Config

type Config struct {
	TMP2PListenAddress string `json:"tmp2plistenaddress"`
	P2PListenAddress   string `json:"p2plistenaddress"`
	RawPrivateKey      string `json:"privatekey"`
	PrivateKey         []byte
	TMPrivateKey       []byte
	SecretConfigPath   string `json:"secretConfigPath"`
	BasePath           string `json:"dataDirectory"`
	IPAddress          string `json:"ipAddress"`
	EthConnection      string `json:"blockchainRPCURL"`
	ContractAddress    string `json:"dkgContractAddress"`
	SelfEpoch          int    `json:"selfEpoch"`
	HttpServerPort     string `json:"port"`
	P2PPort            string `json:"p2pPort"`
	TMP2PPort          string `json:"tmP2PPort"`
	TMRPCPort          string `json:"tmRPCPort"`
	Domain             string `json:"domain"`
	GatewayURL         string `json:"gatewayUrl"`
	PasswordlessUrl    string `json:"passwordlessUrl"`
	OAuthUrl           string `json:"oauthUrl"`
	GlobalKeyCertPool  string `json:"globalKeyCertPool"`
}

func (c *Config) VerifyRequired() error {
	if c.RawPrivateKey == "" && c.SecretConfigPath == "" {
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
	config.TMP2PListenAddress = fmt.Sprintf("tcp://%s:%s", config.IPAddress, config.TMP2PPort)
	config.P2PListenAddress = fmt.Sprintf("/ip4/%s/tcp/%s", config.IPAddress, config.P2PPort)
}

func ReadConfigJson(configPath string) (*Config, error) {
	config := GetDefaultConfig()
	log.Debugf("ConfigPath=%s", configPath)
	f, err := os.OpenFile(configPath, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		log.WithError(err).Error("OpenConfigFile")
		return nil, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(config)
	if err != nil {
		log.WithError(err).Error("DecodeConfig")
		return nil, fmt.Errorf("error reading config: %w", err)
	}
	config.TMP2PListenAddress = fmt.Sprintf("tcp://%s:%s", config.IPAddress, config.TMP2PPort)
	config.P2PListenAddress = fmt.Sprintf("/ip4/%s/tcp/%s", config.IPAddress, config.P2PPort)
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
		PasswordlessUrl:    DefaultPasswordlessUrl,
		OAuthUrl:           DefaultOAuthUrl,
		GlobalKeyCertPool:  DefaultGlobalKeyCertPool,
		P2PPort:            DefaultP2PPort,
		TMP2PPort:          DefaultTmP2PPort,
		TMRPCPort:          DefaultTmRPCPort,
	}
	return config
}

func GetNodePrivateKey(configPath string) (key []byte, err error) {
	return GetSecretFromVault(configPath, secret.NodeKey)
}

func GetTendermintPrivateKey(configPath string) (key []byte, err error) {
	return GetSecretFromVault(configPath, secret.TendermintKey)
}

func GetSecretFromVault(configPath, keyType string) ([]byte, error) {
	c, err := secret.ReadConfig(configPath)
	if err != nil {
		return nil, err
	}

	manager, err := vault.NewVaultManager(c)
	if err != nil {
		return nil, err
	}

	err = manager.Setup()
	if err != nil {
		return nil, err
	}

	key, err := manager.GetSecret(keyType)
	if err != nil {
		return nil, err
	}
	return key, nil
}
