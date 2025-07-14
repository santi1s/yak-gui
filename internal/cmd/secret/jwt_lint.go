package secret

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"

	"github.com/spf13/cobra"
)

func lintJwtSecret(cmd *cobra.Command, args []string) error {
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

	secret, err := ReadSecretData(clients, config.SecretPrefix+"/"+providedFlags.path, providedFlags.version)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`(?m)^\s*{`)
	data := secret.Data["data"].(map[string]any)
	for k, v := range data {
		_v := v.(string)
		if re.MatchString(_v) {
			// check if json is valid
			if valid := json.Valid([]byte(_v)); !valid {
				return fmt.Errorf("invalid JSON in key %s", k)
			}
		}
	}
	cli.Println("No JSON errors found in secret")
	return nil
}

var jwtLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "lint a secret that can contain interservice/json configs",
	Args:  cobra.ExactArgs(0),
	RunE:  lintJwtSecret,
}

func init() {
	jwtLintCmd.Flags().IntVarP(&providedFlags.version, "version", "v", 0, "version of the secret to get")
}
