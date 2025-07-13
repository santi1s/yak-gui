package secret

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SyncInfo struct {
	CI             int
	CIVersions     map[string]interface{}
	Secret         int
	SecretVersions map[string]interface{}
}

func secretCheckSync(cmd *cobra.Command, args []string) error {
	if !providedFlags.allPlatforms && providedFlags.platform == "" {
		return errCheckSyncMissingFlags
	}

	errorFound := false
	platforms := []string{}

	if providedFlags.allPlatforms {
		configPlatforms := viper.GetStringMap("platforms")
		for k := range configPlatforms {
			platforms = append(platforms, k)
		}
	} else {
		platforms = append(platforms, providedFlags.platform)
	}

	if providedFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}

	for _, platform := range platforms {
		environment := "common"
		if platform == "common" {
			environment = ""
		}

		config, err := helper.GetVaultConfig(platform, environment, providedFlags.team)
		if err != nil {
			return err
		}

		clients, err := helper.VaultLoginWithAwsAndGetClients(config)
		if err != nil {
			return err
		}

		allSecretsSyncInfo, err := getNamespaceSyncInfo(clients)
		if err != nil {
			return err
		}

		if len(allSecretsSyncInfo) > 0 {
			errorFound = true

			cli.Printf("Secrets are not synced on %s platform (version -1 means that secret has not been found)\n", platform)
			for k, v := range allSecretsSyncInfo {
				cli.Printf("  %s - secret current version is %d while ci current version is %d\n", k, v.Secret, v.CI)
				cli.Printf("    secret version deletion: ")
				keys := []int{}
				for k1 := range v.SecretVersions {
					bla, _ := strconv.Atoi(k1)
					keys = append(keys, bla)
				}
				sort.Ints(keys)
				for _, k1 := range keys {
					if v.SecretVersions[strconv.Itoa(k1)].(map[string]interface{})["deletion_time"] == "" {
						cli.Printf("%d:false ", k1)
					} else {
						cli.Printf("%d:true ", k1)
					}
				}
				cli.Printf("\n    ci version deletion: ")
				keys = []int{}
				for k1 := range v.CIVersions {
					bla, _ := strconv.Atoi(k1)
					keys = append(keys, bla)
				}
				sort.Ints(keys)
				for _, k1 := range keys {
					if v.CIVersions[strconv.Itoa(k1)].(map[string]interface{})["deletion_time"] == "" {
						cli.Printf("%d:false ", k1)
					} else {
						cli.Printf("%d:true ", k1)
					}
				}
				cli.Printf("\n")
			}
			cli.Printf("\n")
		}
	}

	if errorFound {
		return errSecretsNotSynced
	}

	return nil
}

func getNamespaceSyncInfo(clients []*api.Client) (map[string]*SyncInfo, error) {
	// get all secret paths from the namespace
	paths, err := helper.WalkVaultPath(clients, "")
	if err != nil {
		return nil, err
	}

	// create a map of all paths to check by merging paths representing real secrets and ci ones to do a 2-ways check
	allPathsToCheck := map[string]*SyncInfo{}
	for _, path := range paths {
		if strings.HasPrefix(path, "ci/") {
			path = strings.Replace(path, "ci/", "", 1)
		}

		allPathsToCheck[path] = &SyncInfo{}
	}

	for k, v := range allPathsToCheck {
		// read secret
		s1, err := ReadSecretMetadata(clients, k)
		if err != nil {
			if err == helper.ErrSecretNotFound {
				v.Secret = -1
			} else {
				return nil, err
			}
		} else {
			v.Secret, _ = strconv.Atoi(s1.Data["current_version"].(json.Number).String())
			v.SecretVersions = s1.Data["versions"].(map[string]interface{})
		}

		// read ci
		s2, err := ReadSecretMetadata(clients, "ci/"+k)
		if err != nil {
			if err == helper.ErrSecretNotFound {
				v.CI = -1
			} else {
				return nil, err
			}
		} else {
			v.CI, _ = strconv.Atoi(s2.Data["current_version"].(json.Number).String())
			v.CIVersions = s2.Data["versions"].(map[string]interface{})
		}

		// compare secret and ci current version
		if v.Secret == v.CI {
			deleteFromMap := false
			// check for each version that deletion_time is matching between secret and ci
			for z := range v.SecretVersions {
				if v.SecretVersions[z].(map[string]interface{})["deletion_time"] == "" && v.CIVersions[z].(map[string]interface{})["deletion_time"] == "" {
					deleteFromMap = true
				} else if v.SecretVersions[z].(map[string]interface{})["deletion_time"] != "" && v.CIVersions[z].(map[string]interface{})["deletion_time"] != "" {
					deleteFromMap = true
				} else {
					deleteFromMap = false
				}
			}
			// if everything is equals, we can consider that the secret is synced with ci
			if deleteFromMap {
				delete(allPathsToCheck, k)
			}
		} else {
			// update map with values
			allPathsToCheck[k] = v
		}
	}

	return allPathsToCheck, nil
}

var checkSyncCmd = &cobra.Command{
	Use:   "check-sync",
	Short: "check that secrets are in sync with ci secrets",
	PreRun: func(cmd *cobra.Command, args []string) {
		cmd.Parent().ResetFlags()
		cmd.ResetFlags()
	},
	Args: cobra.ExactArgs(0),
	RunE: secretCheckSync,
}

func init() {
	checkSyncCmd.Flags().StringVarP(&providedFlags.platform, "platform", "P", "", "platform for which secrets are checked")
	checkSyncCmd.Flags().BoolVarP(&providedFlags.allPlatforms, "all-platforms", "A", false, "execute check for all platforms")
	checkSyncCmd.MarkFlagsMutuallyExclusive("platform", "all-platforms")
	err := checkSyncCmd.MarkFlagRequired("platform")
	if err != nil {
		panic(err)
	}
}
