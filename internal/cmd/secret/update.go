package secret

import (
	"os"

	"github.com/doctolib/yak/internal/helper"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

func PatchSecretData(clients []*api.Client, secretPath string, data map[string]interface{}) (*api.Secret, error) {
	const mount = "kv/data/"

	latestVersion, err := GetLatestVersion(clients, secretPath)
	if err != nil {
		return nil, err
	}
	if latestVersion == -1 {
		return nil, helper.ErrSecretNotFound
	}

	currentSecret, err := ReadSecretData(clients, secretPath, latestVersion)
	if err != nil {
		return nil, err
	}

	payload := currentSecret.Data["data"].(map[string]interface{})
	for k, v := range data {
		if v == nil {
			delete(payload, k)
		} else {
			payload[k] = v
		}
	}

	ciPayload := make(map[string]interface{})
	for k := range payload {
		ciPayload[k] = ""
	}

	s, err := writeSecret(clients, mount+secretPath, map[string]interface{}{"data": payload})
	if err != nil {
		return s, err
	}

	s2, err := writeSecret(clients, mount+"ci/"+secretPath, map[string]interface{}{"data": ciPayload})
	if err != nil {
		return s2, err
	}

	return s, nil
}

func secretUpdate(cmd *cobra.Command, args []string) error {
	var input map[string]interface{}
	var err error
	if providedFlags.interactive && len(providedFlags.keys) > 0 {
		return errInteractiveKeysCantBeSetTogether
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
	if len(providedFlags.keys) > 0 {
		if !providedFlags.remove {
			return errKeysCantBeSetWithoutRemove
		}
		input = make(map[string]interface{}, len(providedFlags.keys))
		for _, v := range providedFlags.keys {
			input[v] = nil
		}
	} else if providedFlags.interactive {
		input, err = readFromInteractive(providedFlags.remove)
		if err != nil {
			return err
		}
	} else {
		input, err = readFromStdin()
		if err != nil {
			return err
		}
		if !providedFlags.remove {
			if hasEmptyValue(input) {
				return errCantContainEmptyValue
			}
		} else {
			if hasValue(input) {
				return errCantContainValue
			}
		}
	}

	secretVersion, err := GetLatestVersion(clients, config.SecretPrefix+"/"+providedFlags.path)
	if err != nil {
		return err
	}
	if secretVersion == -1 {
		return helper.ErrSecretNotFound
	}

	data := input
	var secret *api.Secret
	if providedFlags.remove { // remove a value
		dataInput := make(map[string]interface{})
		for k := range input {
			dataInput[k] = nil
		}
		data = dataInput
	}

	secret, err = PatchSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, data)
	if err != nil {
		return err
	}
	return formatOutput(secret.Data)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a secret in vault",
	RunE:  secretUpdate,
}

func init() {
	updateCmd.Flags().BoolVarP(&providedFlags.interactive, "interactive", "i", false, "start an interactive prompt")
	updateCmd.Flags().BoolVar(&providedFlags.remove, "remove", false, "remove the provided keys from the secret")
	updateCmd.Flags().StringSliceVarP(&providedFlags.keys, "keys", "k", []string{}, "comma-separated list of keys to be removed from the secret. Needs --remove flag to be set. Mutually exclusive with --interactive flag.")

	// The following command breaks E2E tests but works for normal behavior, a custom error has been created to handle this
	// updateCmd.MarkFlagsMutuallyExclusive("interactive", "keys")
}
