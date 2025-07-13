package secret

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/doctolib/yak/internal/helper"

	"github.com/doctolib/yak/cli"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

func ReadSecretData(clients []*api.Client, secretPath string, version int) (*api.Secret, error) {
	const mount = "kv/data/"
	return readSecret(clients, mount+secretPath, version)
}

func ReadMatchingSecretDataKeys(clients []*api.Client, secretPath string, version int, key string) (*api.Secret, error) {
	const mount = "kv/data/"
	found := false

	secret, err := readSecret(clients, mount+secretPath, version)
	if err != nil {
		return nil, err
	}

	filteredData := make(map[string]interface{})
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		for k, v := range data {
			if strings.Contains(k, key) {
				filteredData[k] = v
				found = true
			}
		}
	}

	// If no matching keys were found, return errDataKeyNotFound
	if !found {
		return nil, helper.ErrSecretDataKeyNotFound
	}
	secret.Data["data"] = filteredData

	return secret, nil
}

func readSecret(clients []*api.Client, secretPath string, version int) (*api.Secret, error) {
	var secret *api.Secret
	var err error
	client := clients[0] // We only read from the first cluster

	if !strings.HasPrefix(secretPath, "kv/data/") && !strings.HasPrefix(secretPath, "kv/metadata/") {
		secretPath = "kv/data/" + secretPath
	}

	if version == 0 {
		secret, err = client.Logical().Read(secretPath)
	} else {
		data := map[string][]string{
			"version": {strconv.Itoa(version)},
		}

		secret, err = client.Logical().ReadWithData(secretPath, data)
	}

	if err != nil { // permission denied (or else)
		return nil, errors.New("error while reading secret on " + client.Address() + ": " + err.Error())
	}

	if secret == nil { // secret does not exist
		return nil, helper.ErrSecretNotFound
	}

	return secret, nil
}

func secretGet(cmd *cobra.Command, args []string) error {
	if providedFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}
	config, err := helper.GetVaultConfig(providedFlags.platform, providedFlags.environment, providedFlags.team)
	var secret *api.Secret

	if err != nil {
		return err
	}
	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	if providedFlags.dataKey != "" {
		secret, err = ReadMatchingSecretDataKeys(clients, config.SecretPrefix+"/"+providedFlags.path, providedFlags.version, providedFlags.dataKey)
		if err != nil {
			return err
		}
	} else {
		secret, err = ReadSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, providedFlags.version)
		if err != nil {
			return err
		}
	}

	if len(secret.Warnings) > 0 {
		cli.Printf("Got %d warnings:\n", len(secret.Warnings))
		for _, w := range secret.Warnings {
			cli.Printf("  - %s\n", w)
		}
	}

	return formatOutput(secret.Data)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get a secret from vault",
	Args:  cobra.ExactArgs(0),
	RunE:  secretGet,
}

func init() {
	getCmd.Flags().IntVarP(&providedFlags.version, "version", "v", 0, "version of the secret to get")
	getCmd.Flags().StringVarP(&providedFlags.dataKey, "data-key", "K", "", "return any secret data where the specified data-key is a substring of the secret's data key")
}
