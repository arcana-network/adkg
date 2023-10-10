package vault

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/arcana-network/dkgnode/secret"
	vault "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type VaultManager struct {
	serverURL string
	token     string
	namespace string
	client    *vault.Client
}

func NewVaultManager(config *secret.SecretConfig) (*VaultManager, error) {
	if config.ServerURL == "" {
		return nil, errors.New("server url not specified in config")
	}
	if config.Token == "" {
		return nil, errors.New("token not specified in config")
	}
	return &VaultManager{
		serverURL: config.ServerURL,
		token:     config.Token,
		namespace: config.Namespace,
	}, nil
}

func (manager *VaultManager) Setup() error {
	config := vault.DefaultConfig()

	config.Address = manager.serverURL

	client, err := vault.NewClient(config)
	if err != nil {
		return err
	}

	client.SetToken(manager.token)
	log.Infof("namespace=%s", manager.namespace)
	client.SetNamespace(manager.namespace)

	manager.client = client
	return nil
}

func (manager *VaultManager) GetSecret(name string) ([]byte, error) {
	secret, err := manager.client.Logical().Read(fmt.Sprintf("secret/data/%s/%s", manager.namespace, name))
	if err != nil {
		return nil, errors.New("unable to get secret from vault")
	}

	if secret == nil {
		return nil, errors.New("secret not found")
	}

	data, ok := secret.Data["data"]
	if !ok {
		return nil, errors.New("unable to assert data type")
	}

	if data == nil {
		return nil, errors.New("secret not found")
	}

	value, ok := data.(map[string]interface{})[name]
	if !ok {
		return nil, errors.New("secret not found")
	}

	val, ok := value.(string)
	if !ok {
		return nil, errors.New("secret not in string format")
	}

	return hex.DecodeString(val)
}

func (manager *VaultManager) SetSecret(name string, value []byte) error {

	_, err := manager.GetSecret(name)

	if err == nil {
		fmt.Printf("[Warning] %s secret found, overwriting..\n", name)
	}

	data := make(map[string]interface{})
	data[name] = hex.EncodeToString(value)

	_, err = manager.client.Logical().Write(fmt.Sprintf("secret/data/%s/%s", manager.namespace, name), map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("unable to store secret %w", err)
	}

	return nil
}
