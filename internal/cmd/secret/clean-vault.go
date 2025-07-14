package secret

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func secretCleanVault(cmd *cobra.Command, args []string) error {
	if !providedFlags.allPlatforms && providedFlags.platform == "" {
		return errCheckSyncMissingFlags
	}

	if !providedFlags.force {
		cli.Println("dry-run execution. No secret will be destroyed.")
	}

	platforms := []string{}
	if providedFlags.allPlatforms {
		configPlatforms := viper.GetStringMap("platforms")
		for k := range configPlatforms {
			platforms = append(platforms, k)
		}
	} else {
		platforms = append(platforms, providedFlags.platform)
	}

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, platform := range platforms {
		secretsToDelete := []string{}
		deleteErrList := []error{}

		environment := "common"
		if platform == "common" {
			environment = ""
		}

		config, err := helper.GetVaultConfig(platform, environment)
		if err != nil {
			return err
		}

		clients, err := helper.VaultLoginWithAwsAndGetClients(config)
		if err != nil {
			return err
		}

		allPaths, err := helper.WalkVaultPath(clients, "/")
		if err != nil {
			return fmt.Errorf("error while walking through vault path: %w", err)
		}

		paths := []string{}
		for _, path := range allPaths {
			if strings.HasPrefix(path, "ci/") {
				continue
			}

			paths = append(paths, path)
		}

		for _, path := range paths {
			secret, err := ReadSecretMetadata(clients, path)
			if err != nil {
				return err
			}

			if secret.Data["custom_metadata"] == nil {
				return fmt.Errorf("no custom_metadata found for %s", path)
			}

			data := secret.Data["custom_metadata"].(map[string]interface{})
			for k, v := range data {
				if k == "vault_destroy_secret_not_before" {
					date, err := time.Parse("2006-01-02", v.(string))
					if err != nil {
						return err
					}

					if date.Before(currentDate) {
						secretsToDelete = append(secretsToDelete, path)
					}
				}
			}
		}

		if len(secretsToDelete) == 0 {
			cli.Printf("nothing to delete in platform %s", platform)
			continue
		}

		for _, v := range secretsToDelete {
			if providedFlags.force {
				err = destroySecret(clients, v)
				if err != nil {
					deleteErrList = append(deleteErrList, err)
				}
			}

			cli.Printf("destroyed secret path %s in platform %s\n", v, platform)
		}

		if len(deleteErrList) > 0 {
			for _, v := range deleteErrList {
				cli.Println(v)
			}
		}
	}

	return nil
}

func destroySecret(clients []*api.Client, secretPath string) error {
	var err error

	for _, client := range clients {
		_, err = client.Logical().Delete("kv/metadata/" + secretPath)
		if err != nil { // permission denied (or else)
			return errors.New("error while destroying secret " + secretPath + " on " + client.Address() + ": " + err.Error())
		}

		_, err = client.Logical().Delete("kv/metadata/ci/" + secretPath)
		if err != nil { // permission denied (or else)
			return errors.New("error while destroying secret " + secretPath + " (ci) on " + client.Address() + ": " + err.Error())
		}
	}

	return nil
}

var cleanVaultCmd = &cobra.Command{
	Use:   "clean-vault",
	Short: "clean vault secrets that have been marked as destroyed (needs admin rights on vault)",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.ResetFlags()
	},
	Args: cobra.ExactArgs(0),
	RunE: secretCleanVault,
}

func initCleanVaultCmdFlags() {
	cleanVaultCmd.Flags().BoolVar(&providedFlags.force, "force", false, "actually destroy secrets instead of dry-run execution")
	cleanVaultCmd.Flags().StringVarP(&providedFlags.platform, "platform", "P", "", "platform for which secrets are checked")
	cleanVaultCmd.Flags().BoolVarP(&providedFlags.allPlatforms, "all-platforms", "A", false, "execute check for all platforms")
	cleanVaultCmd.MarkFlagsMutuallyExclusive("platform", "all-platforms")
	err := cleanVaultCmd.MarkFlagRequired("platform")
	if err != nil {
		panic(err)
	}
}
func init() {
	initCleanVaultCmdFlags()
}
