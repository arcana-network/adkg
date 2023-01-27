package output

import (
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
		Use:   "output",
		Short: "Used to output public creds",
		Run:   runCommand,
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

	data, err := secret.GetResult(manager)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintln(os.Stdout, data.GetOutput())
}
