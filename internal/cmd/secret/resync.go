package secret

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/doctolib/yak/cli"
	"github.com/spf13/cobra"
)

// Return a bool of wheter a secret is in deleted state or not.
func isSecretDeleted(s *api.Secret) bool {
	return s.Data["metadata"].(map[string]interface{})["deletion_time"] != ""
}

// Return the version of a specified secret
func getSecretVersion(s *api.Secret) (int, error) {
	version, err := strconv.Atoi(s.Data["metadata"].(map[string]interface{})["version"].(json.Number).String())
	if err != nil {
		return 0, err
	}
	return version, nil
}

// Return a bool of whether a secret version exists or not and the secret if it exists
func secretVersionExists(clients []*api.Client, secretPath string, version int) (*api.Secret, bool, error) {
	secret, err := ReadSecretData(clients, secretPath, version)
	if err != nil && err != helper.ErrSecretNotFound {
		// Any other error than "Secret not found"
		return nil, false, err
	} else if err == nil {
		return secret, true, nil
	}
	return nil, false, nil
}

// TODO: merge with deleteSecret by removing double writes on the latter
// Delete a secret version
func secretDeleteVersion(clients []*api.Client, secretPath string, version int) error {
	data := map[string]interface{}{
		"versions": strconv.Itoa(version),
	}
	for _, client := range clients {
		_, err := client.Logical().Write("kv/delete/"+secretPath, data)
		if err != nil { // permission denied (or else)
			return errors.New("error while deleting secret " + secretPath + " on " + client.Address() + ": " + err.Error())
		}
	}
	return nil
}

// Undelete a secret version
func secretUndeleteVersion(clients []*api.Client, secretPath string, version int) error {
	data := map[string]interface{}{
		"versions": strconv.Itoa(version),
	}
	for _, client := range clients {
		_, err := client.Logical().Write("kv/undelete/"+secretPath, data)
		if err != nil { // permission denied (or else)
			return errors.New("error while undeleting secret " + secretPath + " on " + client.Address() + ": " + err.Error())
		}
	}
	return nil
}

// Create a CI secret from a real secret
// Check that latest version is target version minus one
// Write the CI secret removing all values from the kv pairs of the original secret
func duplicateSecretVersionKeys(clients []*api.Client, secretPath string, version int, data map[string]interface{}) error {
	// Making sure the "current_version" counter equals to version -1, otherwise we might end up creating the wrong version
	ciSecret, err := ReadSecretMetadata(clients, secretPath)
	if err != nil && err != helper.ErrSecretNotFound {
		return err
	}
	var currentVersion int
	if err == helper.ErrSecretNotFound {
		currentVersion = 0
	} else {
		currentVersion, err = strconv.Atoi(ciSecret.Data["current_version"].(json.Number).String())
		if err != nil {
			return err
		}
	}

	if currentVersion != version-1 {
		return fmt.Errorf("unexpected version %d for CI secret, expected: %d", currentVersion, version-1)
	}

	// We write the CI secret with the keys from the real one (and empty string as a value)
	var payloadData = make(map[string]interface{})
	for k := range data {
		payloadData[k] = ""
	}
	payload := map[string]interface{}{
		"data": payloadData,
	}
	_, err = writeSecret(clients, "kv/data/"+secretPath, payload)
	if err != nil {
		return err
	}
	return nil
}

// Recreate a CI secret from a real secret with the same version, the same keys (without values) and same status (deleted or not)
// Fail when CI secret is too far behind
func secretResync(cmd *cobra.Command, args []string) error {
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
	secretPath := config.SecretPrefix + "/" + providedFlags.path

	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	// Read real secret
	secret, err := ReadSecretData(clients, secretPath, providedFlags.version)
	if err != nil {
		return err
	}

	if isSecretDeleted(secret) {
		version, err := getSecretVersion(secret)
		if err != nil {
			return err
		}

		if ciSecret, exists, err := secretVersionExists(clients, "ci/"+secretPath, version); err != nil {
			return err
		} else if exists {
			if isSecretDeleted(ciSecret) {
				cli.Println("Nothing to do, CI secret version already deleted")
				return nil
			}
			err = secretDeleteVersion(clients, "ci/"+secretPath, version)
			if err != nil {
				return err
			}
			cli.Println("CI secret succesfully resynchronized (deleted)")
		} else {
			// Create it with blank key/value
			err := duplicateSecretVersionKeys(clients, "ci/"+secretPath, version, map[string]interface{}{})
			if err != nil {
				return err
			}

			// Delete it
			err = secretDeleteVersion(clients, "ci/"+secretPath, version)
			if err != nil {
				return err
			}
			cli.Println("CI secret succesfully resynchronized (created & deleted)")
		}
	} else {
		version, err := getSecretVersion(secret)
		if err != nil {
			return err
		}

		if ciSecret, exists, err := secretVersionExists(clients, "ci/"+secretPath, version); err != nil {
			return err
		} else if exists {
			if isSecretDeleted(ciSecret) {
				// Undelete it
				err = secretUndeleteVersion(clients, "ci/"+secretPath, version)
				if err != nil {
					return err
				}
				cli.Println("CI secret succesfully resynchronized (undeleted)")
			} else {
				cli.Println("Nothing to do, CI secret already exists with this version")
			}
		} else {
			// Create it
			err := duplicateSecretVersionKeys(clients, "ci/"+secretPath, version, secret.Data["data"].(map[string]interface{}))
			if err != nil {
				return err
			}
			cli.Println("CI secret succesfully resynchronized")
		}
	}
	return nil
}

var resyncCmd = &cobra.Command{
	Use:   "resync",
	Short: "resync a secret with its ci secret counterpart",
	Args:  cobra.ExactArgs(0),
	RunE:  secretResync,
}

func init() {
	resyncCmd.Flags().IntVarP(&providedFlags.version, "version", "v", 0, "version of the secret to get")
}
