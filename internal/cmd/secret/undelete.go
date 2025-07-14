package secret

import (
	"errors"
	"os"
	"strconv"

	"github.com/santi1s/yak/internal/helper"

	"github.com/santi1s/yak/cli"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

func undeleteSecret(clients []*api.Client, secretPath string, version int) error {
	var err error

	data := map[string]interface{}{
		"versions": strconv.Itoa(version),
	}
	for _, client := range clients {
		_, err = client.Logical().Write("kv/undelete/"+secretPath, data)
		if err != nil { // permission denied (or else)
			return errors.New("error while undeleting secret on " + client.Address() + ": " + err.Error())
		}

		_, err = client.Logical().Write("kv/undelete/ci/"+secretPath, data)
		if err != nil { // permission denied (or else)
			return errors.New("error while undeleting secret (ci) on " + client.Address() + ": " + err.Error())
		}
	}

	return nil
}

func secretUndelete(cmd *cobra.Command, args []string) error {
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
			cli.Printf("You are about to undelete the version %d of secret [platform:%s,path:%s]\n", providedFlags.version, platform, providedFlags.path)
		} else {
			var environment string
			if providedFlags.environment == "" {
				environment = "common"
			} else {
				environment = providedFlags.environment
			}
			cli.Printf("You are about to undelete the version %d of secret [platform:%s,environment:%s,path:%s]\n", providedFlags.version, providedFlags.platform, environment, providedFlags.path)
		}
	}

	cli.SetSkipConfirmation(providedFlags.skipConfirm)
	if cli.AskConfirmation("Do you want to confirm this action?") {
		clients, err := helper.VaultLoginWithAwsAndGetClients(config)
		if err != nil {
			return err
		}

		err = undeleteSecret(clients, config.SecretPrefix+"/"+providedFlags.path, providedFlags.version)
		if err != nil {
			return err
		}

		cli.Println("secret version has been undeleted (if it existed)")
		return nil
	}

	return errAskConfirmationNotConfirmed
}

var undeleteCmd = &cobra.Command{
	Use:   "undelete",
	Short: "undelete a secret from vault",
	RunE:  secretUndelete,
}

func init() {
	undeleteCmd.Flags().IntVarP(&providedFlags.version, "version", "v", 0, "version of the secret to undelete")
	undeleteCmd.Flags().BoolVar(&providedFlags.skipConfirm, "skip-confirm", false, "will skip confirmation for the undeletion")

	err := undeleteCmd.MarkFlagRequired("version")
	if err != nil {
		panic(err)
	}
}
