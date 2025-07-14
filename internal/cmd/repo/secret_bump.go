package repo

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
)

var errNoReferenceFilesFound = errors.New("no logical secret resources found for this platform")

func DoSecretBump(secretFlags *RepoSecretFlags, cwd string) error {
	config, err := helper.GetVaultConfig(secretFlags.Platform, secretFlags.Environment, secretFlags.team)
	if err != nil {
		return err
	}
	platform := strings.Replace(config.VaultNamespace, "doctolib/", "", 1)

	var ymlFilePath string
	if platform == "common" {
		if secretFlags.SecretProd {
			ymlFilePath = cwd + "/" + secretFolder + "/" + platform + "-prod/*.yml"
		} else {
			ymlFilePath = cwd + "/" + secretFolder + "/" + platform + "-shared/*.yml"
		}
	} else {
		ymlFilePath = cwd + "/" + secretFolder + "/" + platform + "/*.yml"
	}

	ymlReferenceFiles, err := filepath.Glob(ymlFilePath)
	if err != nil {
		return err
	}

	if len(ymlReferenceFiles) == 0 {
		return errNoReferenceFilesFound
	}

	clients, err := helper.VaultLoginWithAwsAndGetClients(config)
	if err != nil {
		return err
	}

	data := map[string][]string{
		"version": {strconv.Itoa(secretFlags.Version)},
	}

	secret, err := clients[0].Logical().ReadWithData("kv/data/"+config.SecretPrefix+"/"+secretFlags.Path, data)
	if err != nil {
		return errors.New("error while reading secret on " + clients[0].Address() + ": " + err.Error())
	}

	if secret == nil { // secret does not exist
		return helper.ErrSecretNotFound
	}

	for _, f := range ymlReferenceFiles {
		logicalSecret := &LogicalSecret{}
		err = UnmarshalRefFile(f, logicalSecret)
		if err != nil {
			return err
		}

		if secret, found := logicalSecret.Secrets[config.SecretPrefix+"/"+secretFlags.Path]; found {
			secret.Version = secretFlags.Version
			logicalSecret.Secrets[config.SecretPrefix+"/"+secretFlags.Path] = secret

			err = writeYmlFile(*logicalSecret, f)
			if err != nil {
				return err
			}

			cli.Printf("bumped version of secret in %s\n", f)
		}
	}

	return nil
}

func secretBump(cmd *cobra.Command, args []string) error {
	err := checkRepositoryPath()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
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
	return DoSecretBump(&providedRepoSecretFlags, cwd)
}

var secretBumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "bump version of a secret in all logical secrets resources in kube repository",
	RunE:  secretBump,
	PreRun: func(cmd *cobra.Command, args []string) {
		err := cmd.Parent().MarkPersistentFlagRequired("path")
		if err != nil {
			panic(err)
		}

		err = cmd.Parent().MarkPersistentFlagRequired("version")
		if err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(0),
}

func init() {
	secretBumpCmd.Flags().BoolVar(&providedRepoSecretFlags.SecretProd, "prod", false, "This flag takes effect only if platform is common. If this flag is true, the secret will be bumped in /common-prod. If the flag is false, the secret will be bumped in /common-shared")
}
