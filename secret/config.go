package secret

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ryanuber/columnize"
)

type SecretConfig struct {
	Kind      string `json:"kind"`
	Token     string `json:"token"`
	Namespace string `json:"namespace"`
	ServerURL string `json:"server_url"`
}

func ReadConfig(path string) (*SecretConfig, error) {
	c, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &SecretConfig{}

	err = json.Unmarshal(c, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *SecretConfig) WriteConfig(path string) error {
	jsonBytes, _ := json.MarshalIndent(c, "", " ")

	return os.WriteFile(path, jsonBytes, 0600)
}

type Result struct {
	Address       string `json:"address"`
	NodeKey       string `json:"nodeKey"`
	TendermintKey string `json:"tendermintkey"`
}

func (r *Result) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[SECRET OUTPUT]\n")
	buffer.WriteString(FormatKV([]string{
		fmt.Sprintf("Node Public Key|%s", r.NodeKey),
		fmt.Sprintf("Node Address(To be shared with Arcana)|%s", r.Address),
		"",
		fmt.Sprintf("Tendermint Public Key|%s", r.TendermintKey),
	}))
	buffer.WriteString("\n")

	return buffer.String()
}

func FormatKV(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = ""
	columnConf.Glue = " = "

	return columnize.Format(in, columnConf)
}
