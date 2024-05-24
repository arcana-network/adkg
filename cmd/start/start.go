package start

import (
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/arcana-network/dkgnode/common"
	"github.com/arcana-network/dkgnode/config"
	"github.com/arcana-network/dkgnode/node"
	"github.com/arcana-network/dkgnode/telemetry"
)

const (
	configFileFlag         = "config"
	blockchainRPCURLFlag   = "rpc-url"
	secretConfigPathFlag   = "secret-config"
	dkgContractAddressFlag = "contract-address"
	gatewayURLFlag         = "gateway-url"
	dataDirFlag            = "data-dir"
	serverPortFlag         = "server-port"
	ipAddressFlag          = "ip-address"
	domainFlag             = "domain"

	FlagMissingError = "required flag missing: %q"
	ConfMissingError = "required config value missing: %q"
)

var cfgFilePath string
var conf = config.GetDefaultConfig()

func GetCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Command to start the node",
		RunE:  runCommand,
	}

	setFlags(cmd)

	// cmd.MarkFlagsRequiredTogether(secretConfigPathFlag, ipAddressFlag)
	// cmd.MarkFlagsMutuallyExclusive(configFileFlag, ipAddressFlag)

	return cmd
}

func setFlags(cmd *cobra.Command) {
	setConfigFileFlags(cmd)
	setParamsFlags(cmd)
}

func setConfigFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&cfgFilePath,
		configFileFlag,
		"./config.json",
		"Used to specify JSON config file path",
	)
}

func setParamsFlags(cmd *cobra.Command) {

	d := config.GetDefaultConfig()
	cmd.Flags().StringVar(
		&conf.SecretConfigPath,
		secretConfigPathFlag,
		"",
		"Used to specify the secret config path",
	)
	cmd.Flags().StringVar(
		&conf.EthConnection,
		blockchainRPCURLFlag,
		d.EthConnection,
		"Used to specify the blockchain url",
	)

	cmd.Flags().StringVar(
		&conf.ContractAddress,
		dkgContractAddressFlag,
		d.ContractAddress,
		"Used to specify the address of the DKG contract",
	)

	cmd.Flags().StringVar(
		&conf.GatewayURL,
		gatewayURLFlag,
		d.GatewayURL,
		"Used to specify the URL for Arcana Gateway",
	)

	cmd.Flags().StringVar(
		&conf.BasePath,
		dataDirFlag,
		"/tmp/keygen-data",
		"Used to specify the data directory used for storing DKG data. Default: '/tmp/keygen-data'",
	)

	cmd.Flags().StringVar(
		&conf.HttpServerPort,
		serverPortFlag,
		"80",
		"Used to specify the server port. Default: '80'",
	)

	cmd.Flags().StringVar(
		&conf.IPAddress,
		ipAddressFlag,
		"",
		"Used to specify the ip address of the node.",
	)

	cmd.Flags().StringVar(
		&conf.Domain,
		domainFlag,
		"",
		"Used to specify the domain name of the current node",
	)

}

func runCommand(cmd *cobra.Command, _ []string) error {
	if common.DoesFileExist(cfgFilePath) {
		c, err := config.ReadConfigJson(cfgFilePath)
		if err != nil {
			log.Infof("Config file parsing error")
			return err
		}
		err = c.VerifyRequired()
		if err != nil {
			log.Infof("Config missing error")
			return err
		}
		conf = c
	} else {
		err := VerifyConfigFromFlags(conf)
		if err != nil {
			log.Infof("Params file flag error %s", err)
			return err
		}
		config.UseIPAdressInListenAddress(conf)
	}

	if conf.RawPrivateKey == "" {
		privateKey, err := config.GetNodePrivateKey(conf.SecretConfigPath)
		if err != nil {
			return err
		}
		conf.PrivateKey = privateKey
		tendermintKey, err := config.GetTendermintPrivateKey(conf.SecretConfigPath)
		if err != nil {
			return err
		}
		conf.TMPrivateKey = tendermintKey
	} else {
		pk, err := hex.DecodeString(conf.RawPrivateKey)
		if err != nil {
			return err
		}
		conf.PrivateKey = pk
	}

	// log.Infof("config: %v", conf)
	go telemetry.StartClient(conf.TelemetryPort)
	node.Start(conf)
	return nil
}

func VerifyConfigFromFlags(conf *config.Config) error {
	if conf.RawPrivateKey == "" && conf.SecretConfigPath == "" {
		return fmt.Errorf(FlagMissingError, secretConfigPathFlag)
	}
	if conf.IPAddress == "" {
		return fmt.Errorf(FlagMissingError, ipAddressFlag)
	}
	return nil
}
