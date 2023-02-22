package root

import (
	"github.com/arcana-network/dkgnode/cmd/secret"
	"github.com/arcana-network/dkgnode/cmd/start"
	"github.com/arcana-network/dkgnode/cmd/version"
	"github.com/spf13/cobra"
)

func GetRootCmd() *cobra.Command {

	var rootCmd = &cobra.Command{}
	rootCmd.AddCommand(start.GetCommand())
	rootCmd.AddCommand(secret.GetCommand())
	rootCmd.AddCommand(version.GetCommand())
	return rootCmd
}
