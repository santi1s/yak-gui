package secret

import (
	"errors"
	"os"
	"strings"

	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"

	"github.com/spf13/cobra"
)

func listSecret(clients []*api.Client, secretPath string) (*api.Secret, error) {
	var secret *api.Secret
	var err error
	client := clients[0] // We only read from the first cluster

	if !strings.HasPrefix(secretPath, "kv/metadata/") {
		secretPath = "kv/metadata/" + secretPath
	}

	secret, err = client.Logical().List(secretPath)
	if err != nil { // permission denied (or else)
		return nil, errors.New("error while listing secret on " + client.Address() + ": " + err.Error())
	}

	if secret == nil { // secret does not exist
		// try to read secret metadata to check if user is putting a secret as path
		_, err = readSecret(clients, secretPath, 0)

		if err == helper.ErrSecretNotFound {
			return nil, helper.ErrSecretPathNotFound
		} else if err != nil { // permission denied (or else)
			return nil, errors.New("error while determining if provided path is a secret or not on " + client.Address() + ": " + err.Error())
		}

		return nil, helper.ErrListSecretNotSecretPath
	}

	return secret, nil
}

func secretList(cmd *cobra.Command, args []string) error {
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

	secret, err := listSecret(clients, config.SecretPrefix+"/"+providedFlags.path)
	if err != nil {
		return err
	}

	return formatOutput(secret.Data)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list secrets from vault",
	Args:  cobra.ExactArgs(0),
	RunE:  secretList,
}

func init() {
	listCmd.Flags().StringVarP(&providedFlags.path, "path", "p", "/", "secret path to list")
}
