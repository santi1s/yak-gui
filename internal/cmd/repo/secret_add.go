package repo

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
)

var (
	errVaultRoleEmpty                  = errors.New("vault role must not be empty as the logical secret resource does not exist yet")
	errVaultRoleNotEmpty               = errors.New("cannot overwrite vault role with add command as the logical secret resource already exists")
	errSecretAlreadyExists             = errors.New("secret already exists in logical secret resource")
	errKeysNotProvided                 = errors.New("secret keys must be set either with keys or all-keys flag")
	errKeyNotFound                     = errors.New("provided key does not exist for this secret version")
	errAllKeysAndKeysCantBeSetTogether = errors.New("all-keys and keys cannot be set together")
)

func secretAdd(cmd *cobra.Command, args []string) error {
	if providedRepoSecretFlags.allKeys && len(providedRepoSecretFlags.keys) > 0 {
		return errAllKeysAndKeysCantBeSetTogether
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

	logicalSecret := &LogicalSecret{Name: providedRepoSecretFlags.name}
	platform := strings.Replace(config.VaultNamespace, "doctolib/", "", 1)

	ymlFilePath, err := getYmlFilePath(platform)
	if err != nil {
		return err
	}

	_, err = os.Stat(ymlFilePath)
	if os.IsNotExist(err) {
		if providedRepoSecretFlags.vaultRole == "" {
			return errVaultRoleEmpty
		}
		// Validate format of JWT subjects
		for _, subject := range providedRepoSecretFlags.jwtSubjects {
			if err := validateTfeJwtSubjectFormat(subject); err != nil {
				return fmt.Errorf("invalid format for TFE JWT subject '%s': %w", subject, err)
			}
		}

		logicalSecret.VaultRole = providedRepoSecretFlags.vaultRole
		logicalSecret.VaultNamespace = config.VaultNamespace
		logicalSecret.TfeJwtSubjects = providedRepoSecretFlags.jwtSubjects
		sort.Strings(logicalSecret.TfeJwtSubjects)
		logicalSecret.Secrets = map[string]Secret{}
	} else if err != nil {
		return err
	} else {
		if providedRepoSecretFlags.vaultRole != "" {
			return errVaultRoleNotEmpty
		}

		err = UnmarshalRefFile(ymlFilePath, logicalSecret)
		if err != nil {
			return err
		}
	}

	if _, ok := logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path]; ok {
		return errSecretAlreadyExists
	}

	if !providedRepoSecretFlags.allKeys && len(providedRepoSecretFlags.keys) == 0 {
		return errKeysNotProvided
	}

	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	data := map[string][]string{
		"version": {strconv.Itoa(providedRepoSecretFlags.Version)},
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

	logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path] = Secret{
		Keys:    keys,
		Version: providedRepoSecretFlags.Version,
	}

	err = writeYmlFile(*logicalSecret, ymlFilePath)
	if err != nil {
		return err
	}

	cli.Printf("secret added to %s\n", ymlFilePath)
	return nil
}

var secretAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a logical secret resource in kube repository",
	RunE:  secretAdd,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := cmd.Parent().MarkPersistentFlagRequired("path")
		if err != nil {
			panic(err)
		}

		err = cmd.Parent().MarkPersistentFlagRequired("version")
		if err != nil {
			panic(err)
		}

		err = cmd.MarkFlagRequired("name")
		if err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(0),
}

func init() {
	secretAddCmd.Flags().StringVarP(&providedRepoSecretFlags.name, "name", "n", "", "name of the logical secret resource")
	secretAddCmd.Flags().StringVarP(&providedRepoSecretFlags.vaultRole, "vault-role", "r", "", "vault role that will be used to connect to Vault cluster")
	secretAddCmd.Flags().StringSliceVar(&providedRepoSecretFlags.jwtSubjects, "tfe-jwt-subjects", []string{}, "JWT subjects for Terraform Enterprise dynamic credentials (format: organization:ORG:project:PROJ:workspace:WS)")
	secretAddCmd.Flags().StringSliceVarP(&providedRepoSecretFlags.keys, "keys", "k", []string{}, "import specified keys from the secret path in the logical secret resource")
	secretAddCmd.Flags().BoolVar(&providedRepoSecretFlags.allKeys, "all-keys", false, "import automatically all keys from the secret path in the logical secret resource")
	secretAddCmd.Flags().BoolVar(&providedRepoSecretFlags.SecretProd, "prod", false, "This flag takes effect only if platform is common. If this flag is true, the secret will be added in /common-prod. If the flag is false, the secret will be added in /common-shared")
}
