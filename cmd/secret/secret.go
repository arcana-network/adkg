package secret

import (
	"bytes"
	"fmt"

	secretGenerate "github.com/arcana-network/dkgnode/cmd/secret/generate"
	secretInit "github.com/arcana-network/dkgnode/cmd/secret/init"
	secretOutput "github.com/arcana-network/dkgnode/cmd/secret/output"
	secretRecover "github.com/arcana-network/dkgnode/cmd/secret/recover"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Command to manage secrets",
	}

	cmd.AddCommand(secretInit.GetCommand())
	cmd.AddCommand(secretGenerate.GetCommand())
	cmd.AddCommand(secretOutput.GetCommand())
	cmd.AddCommand(secretRecover.GetCommand())
	return cmd
}

type Result struct {
	Address       string `json:"address"`
	NodeKey       string `json:"nodeKey"`
	TendermintKey string `json:"tendermintkey"`
}

func (r *Result) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[SECRET INIT]\n")
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
