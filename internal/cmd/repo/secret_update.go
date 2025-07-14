package repo

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
)

var errLogicalSecretResourceDoesNotExist = errors.New("provided logical secret resource name does not exist")
var errVaultRoleAndOtherFlagsCantBeSetTogether = errors.New("vault-role and [environment, path, version, keys, all-keys] can't be set together")
var errSecretDoesNotExist = errors.New("secret does not exist in logical secret resource")

// updateTfeJwtSubjects manages adding and removing TFE JWT subjects
func updateTfeJwtSubjects(current []string, add []string, remove []string) ([]string, error) {
	// Validate format of subjects to add
	for _, subject := range add {
		if err := validateTfeJwtSubjectFormat(subject); err != nil {
			return nil, fmt.Errorf("invalid format for TFE JWT subject '%s': %w", subject, err)
		}
	}

	// Check that subjects to add don't already exist in current list
	for _, toAdd := range add {
		if slices.Contains(current, toAdd) {
			return nil, fmt.Errorf("TFE JWT subject '%s' already exists in current list", toAdd)
		}
	}

	// Check that subjects to remove exist in current list
	for _, toRemove := range remove {
		if !slices.Contains(current, toRemove) {
			return nil, fmt.Errorf("TFE JWT subject '%s' not found in current list: %w", toRemove, errTfeJwtSubjectNotFound)
		}
	}

	// Start with current list
	result := make([]string, len(current))
	copy(result, current)

	// Add new subjects
	result = append(result, add...)

	// Remove subjects
	if len(remove) > 0 {
		filtered := []string{}
		for _, existing := range result {
			if !slices.Contains(remove, existing) {
				filtered = append(filtered, existing)
			}
		}
		result = filtered
	}

	// Sort the result
	sort.Strings(result)
	return result, nil
}

func secretUpdate(cmd *cobra.Command, args []string) error {
	if providedRepoSecretFlags.allKeys && len(providedRepoSecretFlags.keys) > 0 {
		return errAllKeysAndKeysCantBeSetTogether
	}

	if providedRepoSecretFlags.vaultRole != "" &&
		(providedRepoSecretFlags.Environment != "" || providedRepoSecretFlags.Path != "" || providedRepoSecretFlags.Version != 0 || len(providedRepoSecretFlags.keys) > 0 || providedRepoSecretFlags.allKeys) {
		return errVaultRoleAndOtherFlagsCantBeSetTogether
	}

	err := checkRepositoryPath()
	if err != nil {
		return err
	}

	if providedRepoSecretFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedRepoSecretFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}

	config, err := helper.GetVaultConfig(providedRepoSecretFlags.Platform, providedRepoSecretFlags.Environment, providedRepoSecretFlags.team)
	if err != nil {
		return err
	}

	logicalSecret := &LogicalSecret{}
	platform := strings.Replace(config.VaultNamespace, "doctolib/", "", 1)

	ymlFilePath, err := getYmlFilePath(platform)
	if err != nil {
		return err
	}

	_, err = os.Stat(ymlFilePath)
	if os.IsNotExist(err) {
		return errLogicalSecretResourceDoesNotExist
	} else if err != nil {
		return err
	}

	err = UnmarshalRefFile(ymlFilePath, logicalSecret)
	if err != nil {
		return err
	}

	if providedRepoSecretFlags.vaultRole != "" {
		logicalSecret.VaultRole = providedRepoSecretFlags.vaultRole
	}

	if len(providedRepoSecretFlags.addJwtSubjects) > 0 || len(providedRepoSecretFlags.removeJwtSubjects) > 0 {
		updatedTfeJwtSubjects, err := updateTfeJwtSubjects(
			logicalSecret.TfeJwtSubjects,
			providedRepoSecretFlags.addJwtSubjects,
			providedRepoSecretFlags.removeJwtSubjects,
		)
		if err != nil {
			return err
		}
		logicalSecret.TfeJwtSubjects = updatedTfeJwtSubjects
	}

	// Only process secret data if we're not just updating metadata
	hasSecretDataFlags := providedRepoSecretFlags.Path != "" || providedRepoSecretFlags.Version != 0 || len(providedRepoSecretFlags.keys) > 0 || providedRepoSecretFlags.allKeys
	if hasSecretDataFlags {
		if _, ok := logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path]; !ok {
			return errSecretDoesNotExist
		}

		updatedSecret := Secret{}
		if providedRepoSecretFlags.Version != 0 {
			updatedSecret.Version = providedRepoSecretFlags.Version
		} else {
			updatedSecret.Version = logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path].Version
		}

		clients, err := helper.VaultLoginWithAwsAndGetClients(config)
		if err != nil {
			return err
		}

		data := map[string][]string{
			"version": {strconv.Itoa(updatedSecret.Version.(int))},
		}

		secret, err := clients[0].Logical().ReadWithData("kv/data/"+config.SecretPrefix+"/"+providedRepoSecretFlags.Path, data)
		if err != nil {
			return errors.New("error while reading secret on " + clients[0].Address() + ": " + err.Error())
		}

		if secret == nil { // secret does not exist
			return helper.ErrSecretNotFound
		}

		if deletionTime, ok := secret.Data["metadata"].(map[string]interface{})["deletion_time"].(string); ok {
			if deletionTime != "" && secret.Data["data"] == nil {
				return helper.ErrSecretNotFound
			}
		}

		if providedRepoSecretFlags.allKeys || len(providedRepoSecretFlags.keys) > 0 {
			keys := []string{}
			if providedRepoSecretFlags.allKeys {
				for k := range secret.Data["data"].(map[string]interface{}) {
					keys = append(keys, k)
				}
			} else {
				for _, v := range providedRepoSecretFlags.keys {
					if _, exists := secret.Data["data"].(map[string]interface{})[v]; !exists {
						return fmt.Errorf("%v: %s", errKeyNotFound, v)
					}
				}
				keys = providedRepoSecretFlags.keys
			}
			sort.Strings(keys)

			updatedSecret.Keys = keys
		} else {
			updatedSecret.Keys = logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path].Keys
		}

		logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path] = updatedSecret
	}

	err = writeYmlFile(*logicalSecret, ymlFilePath)
	if err != nil {
		return err
	}

	cli.Printf("secret updated in %s\n", ymlFilePath)
	return nil
}

var secretUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a logical secret resource in kube repository",
	RunE:  secretUpdate,
	Args:  cobra.ExactArgs(0),
}

func init() {
	secretUpdateCmd.Flags().StringVarP(&providedRepoSecretFlags.name, "name", "n", "", "name of the logical secret resource")
	secretUpdateCmd.Flags().StringVarP(&providedRepoSecretFlags.vaultRole, "vault-role", "r", "", "vault role that will be used to connect to Vault cluster")
	secretUpdateCmd.Flags().StringSliceVar(&providedRepoSecretFlags.addJwtSubjects, "add-tfe-jwt-subjects", []string{}, "add TFE JWT subjects to existing list (format: organization:ORG:project:PROJ:workspace:WS)")
	secretUpdateCmd.Flags().StringSliceVar(&providedRepoSecretFlags.removeJwtSubjects, "remove-tfe-jwt-subjects", []string{}, "remove TFE JWT subjects from existing list")
	secretUpdateCmd.Flags().StringSliceVarP(&providedRepoSecretFlags.keys, "keys", "k", []string{}, "import specified keys from the secret path in the logical secret resource")
	secretUpdateCmd.Flags().BoolVar(&providedRepoSecretFlags.allKeys, "all-keys", false, "import automatically all keys from the secret path in the logical secret resource")
	secretUpdateCmd.Flags().BoolVar(&providedRepoSecretFlags.SecretProd, "prod", false, "This flag takes effect only if platform is common. If this flag is true, the secret under /common-prod will be updated. If the flag is false, the secret under /common-shared will be updated")

	err := secretUpdateCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}
}
