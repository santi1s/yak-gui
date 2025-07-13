package secret

import (
	"context"
	"errors"
	"os"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

func PatchSecretMetadata(clients []*api.Client, secretPath string, data map[string]interface{}) (*api.Secret, error) {
	const mount = "kv/metadata/"
	var secret *api.Secret
	var err error

	for _, client := range clients {
		secret, err = client.Logical().JSONMergePatch(context.Background(), mount+secretPath, data)
		if err != nil {
			return secret, errors.New("error while patching secret on " + client.Address() + ": " + err.Error())
		}
	}

	return secret, nil
}

func metadataUpdate(cmd *cobra.Command, args []string) error {
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
		return errMetadataKeyNotFound
	}
	if _, ok := secret.Data["custom_metadata"].(map[string]interface{})[providedFlags.key]; ok {
		data := map[string]interface{}{
			"custom_metadata": map[string]interface{}{
				providedFlags.key: providedFlags.value,
			},
		}

		_, err := PatchSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, data)
		if err != nil {
			return err
		}

		cli.Printf("metadata %s has been updated\n", providedFlags.key)
		return nil
	}
	return errMetadataKeyNotFound
}

var metadataUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a secret's metadata from vault",
	Args:  cobra.ExactArgs(0),
	RunE:  metadataUpdate,
}

func init() {
	var err error

	metadataUpdateCmd.Flags().StringVarP(&providedFlags.key, "key", "k", "", "key of the metadata (mandatory)")
	metadataUpdateCmd.Flags().StringVarP(&providedFlags.value, "value", "v", "", "value of the metadata (mandatory)")

	err = metadataUpdateCmd.MarkFlagRequired("key")
	if err != nil {
		panic(err)
	}

	err = metadataUpdateCmd.MarkFlagRequired("value")
	if err != nil {
		panic(err)
	}
}
