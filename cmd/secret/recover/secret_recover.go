package recover

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/secret"
	"github.com/arcana-network/dkgnode/secret/vault"
	"github.com/spf13/cobra"
)

var configPath string
var nodeKey string
var tendermintKey string

const (
	configFlag = "secret-config"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Used to output public creds",
		Run:   runCommand,
	}

	setFlags(cmd)

	_ = cmd.MarkFlagRequired(configFlag)
	_ = cmd.MarkFlagRequired("node-key")
	_ = cmd.MarkFlagRequired("tendermint-key")

	return cmd
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&configPath,
		configFlag,
		"",
		"path to secret config file",
	)
	cmd.Flags().StringVar(
		&nodeKey,
		"node-key",
		"",
		"base64 encoded node key",
	)
	cmd.Flags().StringVar(
		&tendermintKey,
		"tendermint-key",
		"",
		"base64 encoded tendermint key",
	)
}

func runCommand(cmd *cobra.Command, args []string) {
	config, err := secret.ReadConfig(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	manager, err := vault.NewVaultManager(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = manager.Setup()
	if err != nil {
		fmt.Println(err)
		return
	}

	nodeKeyBytes, err := base64.StdEncoding.DecodeString(nodeKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = manager.SetSecret(secret.NodeKey, nodeKeyBytes)
	if err != nil {
		return
	}

	tmKeyBytes, err := base64.StdEncoding.DecodeString(tendermintKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = manager.SetSecret(secret.TendermintKey, tmKeyBytes)
	if err != nil {
		return
	}

	res, err := secret.GetResult(manager)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintln(os.Stdout, res.GetOutput())
}
