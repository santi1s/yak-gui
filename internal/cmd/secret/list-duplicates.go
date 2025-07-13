package secret

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
)

func secretListDuplicates(cmd *cobra.Command, args []string) error {
	secretValue := ""
	if providedFlags.interactive {
		cli.Print("Enter the secret value to check: ")
		interactiveSecretValue, err := cli.ReadPassword()
		if err != nil {
			return fmt.Errorf("error while asking for secret value: %w", err)
		}
		cli.Println()
		secretValue = interactiveSecretValue
	} else {
		input, err := cli.ReadAll()
		if err != nil {
			return err
		}
		secretValue = input
	}

	secretValue = strings.TrimSpace(secretValue)
	if secretValue == "" {
		return errEmptyInput
	}

	platform := providedFlags.platform
	environment := providedFlags.environment
	path := providedFlags.path

	if providedFlags.team != "" {
		awsFeatureTeamConfigFile, err := helper.AddAWSConfigProfileForFeatureTeam(providedFlags.team)
		if err != nil {
			return err
		}
		defer os.Remove(awsFeatureTeamConfigFile)
	}

	config, err := helper.GetVaultConfig(platform, environment, providedFlags.team)
	if err != nil {
		return err
	}

	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	cli.Printf("Walking through path %s ...\n", path)
	paths, err := helper.WalkVaultPath(clients, path)
	if err != nil {
		return fmt.Errorf("error while walking through vault path: %w", err)
	}

	duplicates := make(map[string][]string)
	for _, path := range paths {
		secret, err := ReadSecretData(clients, path, 0)
		if err != nil {
			return err
		}
		secretData := secret.Data["data"]
		if secretData == nil {
			continue
		}
		data := secretData.(map[string]interface{})
		for k, v := range data {
			if v == secretValue {
				duplicates[path] = append(duplicates[path], k)
			}
		}
	}

	if len(duplicates) == 0 {
		cli.Println("Secret not found.")
		return nil
	}

	if len(duplicates) == 1 {
		for path, keys := range duplicates {
			if len(keys) == 1 {
				cli.Printf("No duplicates found for provided secret value. Secret located at %s:%s\n", path, keys[0])
				return nil
			}
		}
		return nil
	}

	// Sort paths, so that the output is deterministic
	allPaths := make([]string, 0)
	for k := range duplicates {
		allPaths = append(allPaths, k)
	}
	sort.Strings(allPaths)

	cli.Println("Duplicates found:")
	for _, p := range allPaths {
		cli.Printf("\t* %s\n", p)
		for _, key := range duplicates[p] {
			cli.Printf("\t\t- %s\n", key)
		}
	}

	return nil
}

var listDuplicatesCmd = &cobra.Command{
	Use:   "list-duplicates",
	Short: "list in the given platform all secrets duplicated with the specified secret value under the specified vault path",
	Args:  cobra.ExactArgs(0),
	RunE:  secretListDuplicates,
}

func init() {
	listDuplicatesCmd.Flags().BoolVarP(&providedFlags.interactive, "interactive", "i", false, "start an interactive prompt")
}
