package secret

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/santi1s/yak/internal/helper"

	"github.com/santi1s/yak/cli"
	"github.com/spf13/cobra"
)

func secretDestroy(cmd *cobra.Command, args []string) error {
	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment, providedFlags.team)
	if err != nil {
		return err
	}

	var environment string
	var platform string

	if providedFlags.platform == "" {
		platform = "common"
	} else {
		platform = providedFlags.platform
	}

	if providedFlags.environment == "" {
		environment = "common"
	} else {
		environment = providedFlags.environment
	}

	oneWeekDate := time.Now().AddDate(0, 0, 8).Format("2006-01-02")
	cli.Printf("You are about to mark the secret [platform:%s,environment:%s,path:%s] to be destroyed in one week (%s).\nAll versions will be deleted.\n", platform, environment, providedFlags.path, oneWeekDate)

	cli.SetSkipConfirmation(false)
	if cli.AskConfirmation("This is a destructive action. Do you want to confirm this action?") {
		if cli.AskConfirmation("Are you really sure of what you are doing?") {
			clients, err := helper.VaultLoginWithAwsAndGetClients(config)
			if err != nil {
				return err
			}

			secret, err := ReadSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path)
			if err != nil {
				return err
			}

			currentVersion, err := strconv.Atoi(secret.Data["current_version"].(json.Number).String())
			if err != nil {
				return err
			}

			oldestVersion, err := strconv.Atoi(secret.Data["oldest_version"].(json.Number).String())
			if err != nil {
				return err
			}

			for i := currentVersion; i > oldestVersion; i-- {
				err = deleteSecret(clients, config.SecretPrefix+"/"+providedFlags.path, i)
				if err != nil {
					return err
				}
			}

			cli.Println("all available versions of the secret have been deleted.")

			data := map[string]interface{}{
				"custom_metadata": map[string]interface{}{
					"vault_destroy_secret_not_before": oneWeekDate,
				},
			}

			_, err = PatchSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, data)
			if err != nil {
				return err
			}

			cli.Println("secret has been marked for destruction.")

			return nil
		}
	}

	return errAskConfirmationNotConfirmed
}

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "mark a secret to be destroyed from vault",
	RunE:  secretDestroy,
}
