package vault

import (
	"errors"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errPlatformMustBeProvided = errors.New("you must provide a platform value")
var errVaultClusterNotFound = errors.New("vault endpoint does not exist in provided configuration file")
var errInvalidRegion = errors.New("invalid region. Only eu-central-1, eu-west-3 and eu-west-1 are allowed")

func getVaultEndpoint(cluster string) (string, error) {
	endpoint := viper.GetString("clusters." + cluster + ".endpoint")

	if endpoint == "" {
		return "", errVaultClusterNotFound
	}

	return endpoint, nil
}

func getToken(cmd *cobra.Command, args []string) error {
	var clients []*api.Client

	var config *helper.VaultConfig
	var err error
	if providedFlags.admin {
		endpoint, err := getVaultEndpoint(providedFlags.cluster)
		if err != nil {
			return err
		}

		switch providedFlags.region {
		case "eu-central-1", "eu-west-3", "eu-west-1":
			break
		default:
			return errInvalidRegion
		}

		config = &helper.VaultConfig{
			Endpoints:      []string{endpoint},
			AwsProfile:     "shared-sso",
			AwsRegion:      providedFlags.region,
			VaultRole:      "Admin",
			VaultNamespace: "",
		}
	} else {
		if providedFlags.platform == "" {
			providedFlags.platform = "common"
		}

		environment := "common"
		if providedFlags.platform == "common" {
			environment = ""
		}

		config, err = helper.GetVaultConfig(providedFlags.platform, environment, providedFlags.team)
		if err != nil {
			return err
		}
	}

	clients, err = helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	result := map[string]string{}
	for _, c := range clients {
		result[c.Address()] = c.Token()
	}

	switch {
	case providedFlags.json:
		return cli.PrintJSON(result)
	case providedFlags.yaml:
		return cli.PrintYAML(result)
	default:
		return cli.PrintYAML(result)
	}
}

var getTokenCmd = &cobra.Command{
	Use:   "get-token",
	Short: "get auth token on a vault cluster",
	RunE:  getToken,
	Example: `Get an admin token for a cluster (<cluster> is one key of the clusters map in your yak secret.yml configuration file):
get-token --admin --cluster <cluster>

Get a token for a given platform (it will return a token for all clusters used by this platform):
get-token --platform dev`,
}

func init() {
	getTokenCmd.Flags().StringVarP(&providedFlags.platform, "platform", "P", "", "platform from which to get a token. Mutually exclusive with --admin and --cluster")
	getTokenCmd.Flags().StringVarP(&providedFlags.region, "region", "r", "eu-central-1", "aws region to use for authentication. Mutually exclusive with --platform")
	getTokenCmd.Flags().StringVar(&providedFlags.cluster, "cluster", "", "cluster from which to get a token. Needs --admin to be set. Mutually exclusive with --platform")
	getTokenCmd.Flags().StringVarP(&providedFlags.team, "team", "t", "", "use this flag only if you're a member of a feature team")
	getTokenCmd.Flags().BoolVar(&providedFlags.admin, "admin", false, "get an admin token. Must be used only as a breaking glass! Needs --cluster to be set. Mutually exclusive with --platform")
	getTokenCmd.MarkFlagsRequiredTogether("admin", "cluster")
	getTokenCmd.MarkFlagsMutuallyExclusive("platform", "admin")
	getTokenCmd.MarkFlagsMutuallyExclusive("platform", "cluster")
	getTokenCmd.MarkFlagsMutuallyExclusive("platform", "region")
	vaultCmd.AddCommand(getTokenCmd)
}
