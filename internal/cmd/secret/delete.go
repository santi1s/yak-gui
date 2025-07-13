package secret

import (
	"errors"
	"os"
	"strconv"

	"github.com/doctolib/yak/internal/helper"

	"github.com/doctolib/yak/cli"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

func deleteSecret(clients []*api.Client, secretPath string, version int) error {
	var err error

	if version > 0 {
		data := map[string]interface{}{
			"versions": strconv.Itoa(version),
		}
		for _, client := range clients {
			_, err = client.Logical().Write("kv/delete/"+secretPath, data)
			if err != nil { // permission denied (or else)
				return errors.New("error while deleting secret on " + client.Address() + ": " + err.Error())
			}

			_, err = client.Logical().Write("kv/delete/ci/"+secretPath, data)
			if err != nil { // permission denied (or else)
				return errors.New("error while deleting secret (ci) on " + client.Address() + ": " + err.Error())
			}
		}
	} else {
		for _, client := range clients {
			_, err = client.Logical().Delete("kv/data/" + secretPath)
			if err != nil { // permission denied (or else)
				return errors.New("error while deleting secret on " + client.Address() + ": " + err.Error())
			}

			_, err = client.Logical().Delete("kv/data/ci/" + secretPath)
			if err != nil { // permission denied (or else)
				return errors.New("error while deleting secret (ci) on " + client.Address() + ": " + err.Error())
			}
		}
	}

	return nil
}

func secretDelete(cmd *cobra.Command, args []string) error {
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

	if !providedFlags.skipConfirm {
		if providedFlags.platform == "" {
			platform := "common"
			if providedFlags.version > 0 {
				cli.Printf("You are about to delete the version %d of secret [platform:%s,path:%s]\n", providedFlags.version, platform, providedFlags.path)
			} else {
				cli.Printf("You are about to delete latest version of secret [platform:%s,path:%s]\n", platform, providedFlags.path)
			}
		} else {
			var environment string
			if providedFlags.environment == "" {
				environment = "common"
			} else {
				environment = providedFlags.environment
			}
			if providedFlags.version > 0 {
				cli.Printf("You are about to delete the version %d of secret [platform:%s,environment:%s,path:%s]\n", providedFlags.version, providedFlags.platform, environment, providedFlags.path)
			} else {
				cli.Printf("You are about to delete latest version of secret [platform:%s,environment:%s,path:%s]\n", providedFlags.platform, environment, providedFlags.path)
			}
		}
	}

	cli.SetSkipConfirmation(providedFlags.skipConfirm)
	if cli.AskConfirmation("Do you want to confirm this action?") {
		clients, err := helper.VaultLoginWithAwsAndGetClients(config)
		if err != nil {
			return err
		}

		err = deleteSecret(clients, config.SecretPrefix+"/"+providedFlags.path, providedFlags.version)
		if err != nil {
			return err
		}

		cli.Println("secret version has been deleted (if it existed)")
		return nil
	}

	return errAskConfirmationNotConfirmed
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a secret from vault",
	RunE:  secretDelete,
}

func init() {
	deleteCmd.Flags().IntVarP(&providedFlags.version, "version", "v", 0, "version of the secret to delete")
	deleteCmd.Flags().BoolVar(&providedFlags.skipConfirm, "skip-confirm", false, "will skip confirmation for the deletion")
}
