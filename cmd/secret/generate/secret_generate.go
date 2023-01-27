package generate

import (
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/secret"
	"github.com/spf13/cobra"
)

const (
	secretConfigFlag = "secret-config"
	tokenFlag        = "token"
	serverURLFlag    = "server-url"
	namespaceFlag    = "namespace"
)

var secretConfig string
var token string
var serverURL string
var namespace string

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Used to generate secret config",
		Run:   runCommand,
	}

	setFlags(cmd)

	_ = cmd.MarkFlagRequired(tokenFlag)
	_ = cmd.MarkFlagRequired(serverURLFlag)

	return cmd
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&secretConfig,
		secretConfigFlag,
		"./secretConfig.json",
		"the path to the secrets manager configuration file",
	)

	cmd.Flags().StringVar(
		&token,
		tokenFlag,
		"",
		"the access token for the secret service",
	)

	cmd.Flags().StringVar(
		&serverURL,
		serverURLFlag,
		"",
		"the server URL for the secret service",
	)
	cmd.Flags().StringVar(
		&namespace,
		namespaceFlag,
		"default",
		"the namespace for the secret service",
	)
}

func runCommand(cmd *cobra.Command, args []string) {
	config := secret.SecretConfig{
		Kind:      "hashicorp-vault",
		Token:     token,
		ServerURL: serverURL,
		Namespace: namespace,
	}

	if err := config.WriteConfig(secretConfig); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Printf("Secret config generated at %s\n", secretConfig)
}
