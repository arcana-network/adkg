package version

import (
	"fmt"
	"os"

	"github.com/arcana-network/dkgnode/versioning"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Command to show current binary version",
		Run:   runCommand,
	}

	return cmd
}

func runCommand(c *cobra.Command, args []string) {
	fmt.Fprintln(os.Stdout, versioning.Version)
}
