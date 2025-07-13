package secret

import (
	"os"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

func ReadSecretMetadata(clients []*api.Client, secretPath string) (*api.Secret, error) {
	const mount = "kv/metadata/"
	return readSecret(clients, mount+secretPath, 0)
}

func metadataGet(cmd *cobra.Command, args []string) error {
	if providedFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}
	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment, providedFlags.team)
	if err != nil {
		return err
	}
	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	secret, err := ReadSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path)
	if err != nil {
		return err
	}

	return formatOutput(secret.Data)
}

var metadataGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get a secret from vault",
	Args:  cobra.ExactArgs(0),
	RunE:  metadataGet,
}
