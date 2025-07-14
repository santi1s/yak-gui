package repo

import (
	"os"
	"strings"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
)

func secretDelete(cmd *cobra.Command, args []string) error {
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
	} else {
		err = UnmarshalRefFile(ymlFilePath, logicalSecret)
		if err != nil {
			return err
		}
	}

	if _, ok := logicalSecret.Secrets[config.SecretPrefix+"/"+providedRepoSecretFlags.Path]; !ok {
		return errSecretDoesNotExist
	}

	delete(logicalSecret.Secrets, config.SecretPrefix+"/"+providedRepoSecretFlags.Path)

	if len(logicalSecret.Secrets) == 0 {
		err = os.Remove(ymlFilePath)
		if err != nil {
			return err
		}

		cli.Printf("secret deleted from %s and file removed as there is no more secret\n", ymlFilePath)
	} else {
		err = writeYmlFile(*logicalSecret, ymlFilePath)
		if err != nil {
			return err
		}

		cli.Printf("secret deleted from %s\n", ymlFilePath)
	}

	return nil
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a logical secret resource in kube repository",
	RunE:  secretDelete,
	Args:  cobra.ExactArgs(0),
}

func init() {
	secretDeleteCmd.Flags().StringVarP(&providedRepoSecretFlags.name, "name", "n", "", "name of the logical secret resource")
	secretDeleteCmd.Flags().BoolVar(&providedRepoSecretFlags.SecretProd, "prod", false, "This flag takes effect only if platform is common. If this flag is true, the secret will be deleted from /common-prod. If the flag is false, the secret will be deleted from /common-shared")

	err := secretDeleteCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}
}
