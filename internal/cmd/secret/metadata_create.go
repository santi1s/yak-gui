package secret

import (
	"os"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/doctolib/yak/cli"
)

// Writes secret metadata
// Returns any error
func WriteSecretMetadata(clients []*api.Client, secretPath string, metadata map[string]interface{}) error {
	payload := map[string]interface{}{
		"custom_metadata": metadata,
	}
	_, err := writeSecret(clients, "kv/metadata/"+secretPath, payload)
	if err != nil {
		return err
	}
	return nil
}

func metadataCreate(cmd *cobra.Command, args []string) error {
	if providedFlags.key == "" {
		return errKeyParameterCantBeEmpty
	} else if providedFlags.value == "" {
		return errValueParameterCantBeEmpty
	}

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

	if secret.Data["custom_metadata"] == nil {
		return errMetadataEmpty
	}
	if _, ok := secret.Data["custom_metadata"].(map[string]interface{})[providedFlags.key]; !ok {
		data := map[string]interface{}{
			"custom_metadata": map[string]interface{}{
				providedFlags.key: providedFlags.value,
			},
		}

		_, err := PatchSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, data)
		if err != nil {
			return err
		}
		cli.Printf("metadata %s has been created\n", providedFlags.key)
		return nil
	}
	return errMetadataAlreadyExists
}

var metadataCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a secret's metadata from vault",
	Args:  cobra.ExactArgs(0),
	RunE:  metadataCreate,
}

func init() {
	var err error

	metadataCreateCmd.Flags().StringVarP(&providedFlags.key, "key", "k", "", "key of the metadata (mandatory)")
	metadataCreateCmd.Flags().StringVarP(&providedFlags.value, "value", "v", "", "value of the metadata (mandatory)")

	err = metadataCreateCmd.MarkFlagRequired("key")
	if err != nil {
		panic(err)
	}

	err = metadataCreateCmd.MarkFlagRequired("value")
	if err != nil {
		panic(err)
	}
}
