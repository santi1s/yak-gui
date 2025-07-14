package secret

import (
	"os"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
)

func metadataDelete(cmd *cobra.Command, args []string) error {
	switch providedFlags.key {
	case "":
		return errKeyParameterCantBeEmpty
	case "owner", "source", "usage":
		return errMetadataNotDeletable
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
	if _, ok := secret.Data["custom_metadata"].(map[string]interface{})[providedFlags.key]; ok {
		data := map[string]interface{}{
			"custom_metadata": map[string]interface{}{
				providedFlags.key: nil,
			},
		}

		if !providedFlags.skipConfirm {
			cli.Printf("You are about to delete the metadata %s from secret [platform:%s,environment:%s,path:%s]\n", providedFlags.key, providedFlags.platform, providedFlags.environment, providedFlags.path)
		}

		cli.SetSkipConfirmation(providedFlags.skipConfirm)
		if cli.AskConfirmation("Do you want to confirm this action?") {
			_, err := PatchSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, data)
			if err != nil {
				return err
			}
			cli.Printf("metadata %s has been deleted\n", providedFlags.key)
			return nil
		}
		return errAskConfirmationNotConfirmed
	}
	return errMetadataKeyNotFound
}

var metadataDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a secret's metadata from vault",
	Args:  cobra.ExactArgs(0),
	RunE:  metadataDelete,
}

func init() {
	var err error

	metadataDeleteCmd.Flags().StringVarP(&providedFlags.key, "key", "k", "", "key of the metadata (mandatory)")
	metadataDeleteCmd.Flags().BoolVar(&providedFlags.skipConfirm, "skip-confirm", false, "will skip confirmation for the deletion")

	err = metadataDeleteCmd.MarkFlagRequired("key")
	if err != nil {
		panic(err)
	}
}
