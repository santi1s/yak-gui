package secret

import (
	"errors"
	"fmt"
	"os"

	"github.com/santi1s/yak/internal/helper"

	"github.com/hashicorp/vault/api"

	"github.com/spf13/cobra"
)

// Writes secret data
// Returns API response and any error
func WriteSecretData(clients []*api.Client, secretPath string, data map[string]interface{}) (*api.Secret, error) {
	payload := map[string]interface{}{
		"data": data,
	}

	ciData := make(map[string]interface{})
	for k := range data {
		ciData[k] = ""
	}

	ciPayload := map[string]interface{}{
		"data": ciData,
	}

	s, err := writeSecret(clients, "kv/data/"+secretPath, payload)
	if err != nil {
		return s, err
	}

	s2, err := writeSecret(clients, "kv/data/ci/"+secretPath, ciPayload)
	if err != nil {
		return s2, err
	}

	return s, nil
}

// Creates either secret data or metadata
// Returns API response and any error
func writeSecret(clients []*api.Client, actualPath string, payload map[string]interface{}) (*api.Secret, error) {
	var s *api.Secret
	var err error
	for _, client := range clients {
		s, err = client.Logical().Write(actualPath, payload)
		if err != nil {
			return s, errors.New("error while writing secret on " + client.Address() + ": " + err.Error())
		}
	}
	// If no error, we return the last secret written (all secrets should be equal)
	return s, nil
}

// Entrypoint of the `secret create` command
// Retrieves a client, checks for non-existence and creates a secret with its metadata
func secretCreate(cmd *cobra.Command, args []string) error {
	var data map[string]interface{}
	var err error

	if providedFlags.owner == "" {
		return errOwnerCantBeEmpty
	} else if providedFlags.source == "" {
		return errSourceCantBeEmpty
	} else if providedFlags.usage == "" {
		return errUsageCantBeEmpty
	}

	metadata := map[string]interface{}{
		"owner":  providedFlags.owner,
		"source": providedFlags.source,
		"usage":  providedFlags.usage,
	}

	if providedFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}

	// create Vault clients
	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment, providedFlags.team)
	if err != nil {
		return err
	}
	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	if providedFlags.interactive {
		data, err = readFromInteractive(false)
		if err != nil {
			return err
		}
	} else {
		data, err = readFromStdin()
		if err != nil {
			return err
		}
		if hasEmptyValue(data) {
			return errCantContainEmptyValue
		}
	}

	// check if the secret exists
	secretVersion, err := GetLatestVersion(clients, config.SecretPrefix+"/"+providedFlags.path)
	if err != nil {
		return err
	}
	if secretVersion != -1 {
		return errSecretAlreadyExists
	}

	secret, err := WriteSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, data)
	if err != nil {
		return err
	}
	err = WriteSecretMetadata(clients, config.SecretPrefix+"/"+providedFlags.path, metadata)
	if err != nil {
		return fmt.Errorf("%s: %s", errMetadataCouldNotBeAdded, err)
	}

	secret.Data["custom_metadata"] = metadata
	return formatOutput(secret.Data)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a secret on vault",
	RunE:  secretCreate,
}

func init() {
	var err error

	createCmd.Flags().BoolVarP(&providedFlags.interactive, "interactive", "i", false, "start an interactive prompt")
	createCmd.Flags().StringVarP(&providedFlags.owner, "owner", "o", "", "owner of the secret (mandatory)")
	createCmd.Flags().StringVarP(&providedFlags.usage, "usage", "u", "", "where is the secret used (mandatory)")
	createCmd.Flags().StringVarP(&providedFlags.source, "source", "s", "", "source of the secret (mandatory)")

	err = createCmd.MarkFlagRequired("owner")
	if err != nil {
		panic(err)
	}

	err = createCmd.MarkFlagRequired("usage")
	if err != nil {
		panic(err)
	}

	err = createCmd.MarkFlagRequired("source")
	if err != nil {
		panic(err)
	}
}
