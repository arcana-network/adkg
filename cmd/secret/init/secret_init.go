package init

import (
	"errors"
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/secret"
	"github.com/arcana-network/dkgnode/secret/vault"
	"github.com/spf13/cobra"
)

var configPath string

const (
	configFlag = "secret-config"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Used to initialize private keys for the node",
		PreRunE: preRunE,
		Run:     runCommand,
	}

	setFlags(cmd)

	_ = cmd.MarkFlagRequired(configFlag)

	return cmd
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&configPath,
		configFlag,
		"",
		"path to secret config file",
	)
}

func preRunE(cmd *cobra.Command, args []string) error {
	if configPath == "" {
		return errors.New("config value not passed")
	}

	return nil
}

func runCommand(cmd *cobra.Command, _ []string) {
	// Init secret, get output
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

	publicKey, address, err := secret.InitNodeKey(manager)
	if err != nil {
		fmt.Println(err)
		return
	}

	tmKey, err := secret.InitTendermintKey(manager)
	if err != nil {
		fmt.Println(err)
		return
	}

	res := secret.Result{Address: address, NodeKey: publicKey, TendermintKey: tmKey}

	fmt.Fprintln(os.Stdout, res.GetOutput())
}
