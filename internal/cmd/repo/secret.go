package repo

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"github.com/doctolib/yak/internal/constant"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type RepoSecretFlags struct {
	cfgFile             string
	checkAll            bool
	checkExistence      bool
	checkVersion        bool
	checkVaultRole      bool
	checkVaultNamespace bool
	checkTfeJwtSubjects bool
	name                string
	vaultRole           string
	jwtSubjects         []string
	addJwtSubjects      []string
	removeJwtSubjects   []string
	keys                []string
	allKeys             bool
	Platform            string
	Environment         string
	team                string
	Path                string
	Version             int
	diff                bool
	check               bool
	SecretProd          bool
	ignoredPrefixes     []string
}

const secretFolder = "configs/vault-secrets" //#nosec

var providedRepoSecretFlags RepoSecretFlags

// TFE JWT subject validation errors
var errInvalidTfeJwtSubjectFormat = errors.New("TFE JWT subject must follow format: organization:ORG:project:PROJECT:workspace:WORKSPACE")
var errTfeJwtSubjectNotFound = errors.New("provided TFE JWT subject does not exist in the list")

// Pre-compiled regex for TFE JWT subject validation to avoid recompilation on each call
var tfeJwtSubjectRegex = regexp.MustCompile(`^organization:[^:]+:project:[^:]+:workspace:[^:]+$`)

// validateTfeJwtSubjectFormat validates that TFE JWT subject follows the required format
// Pattern: organization:ORG:project:PROJECT:workspace:WORKSPACE
// Where ORG, PROJECT, WORKSPACE can be wildcard (*) or actual values
func validateTfeJwtSubjectFormat(subject string) error {
	if !tfeJwtSubjectRegex.MatchString(subject) {
		return errInvalidTfeJwtSubjectFormat
	}
	return nil
}

var initConfig = internalInitConfig
var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "manage a secret in kube or terraform-infra repository",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
			cmd.Root().PersistentPreRun(cmd, args)
		}

		if cmd.HasParent() && cmd.Parent().Name() == "secret" {
			_ = viper.BindPFlag("config", cmd.Parent().PersistentFlags().Lookup("config"))
		}

		initConfig(cmd, args)
	},
}

func checkRepositoryPath() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return err
	}

	kubeRepositoryPath := os.Getenv("KUBE_REPOSITORY_PATH")
	kubeRepositoryPath, err = filepath.Abs(kubeRepositoryPath)
	if err != nil {
		return err
	}

	tfinfraRepositoryPath := os.Getenv("TFINFRA_REPOSITORY_PATH")
	tfinfraRepositoryPath, err = filepath.Abs(tfinfraRepositoryPath)
	if err != nil {
		return err
	}

	if cwd == kubeRepositoryPath || cwd == tfinfraRepositoryPath {
		_, err = os.Stat(cwd + "/configs/vault-secrets")
		if err != nil {
			return err
		}

		return nil
	}

	return errNotInRepoFolder
}

func EncodeLogicalSecretToYmlReferenceFile(logicalSecret LogicalSecret, buffer *bytes.Buffer) error {
	ymlReferenceFile := &YmlReferenceFile{
		LogicalSecret: map[string]LogicalSecret{
			logicalSecret.Name: logicalSecret,
		},
	}
	buffer.Write([]byte(`---
`))
	encoder := yaml.NewEncoder(buffer)
	encoder.SetIndent(2)
	return encoder.Encode(ymlReferenceFile)
}

func writeYmlFile(logicalSecret LogicalSecret, path string) error {
	content := bytes.Buffer{}
	err := EncodeLogicalSecretToYmlReferenceFile(logicalSecret, &content)
	if err != nil {
		return err
	}

	if _, err = os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(path), 0755) //#nosec
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(path, content.Bytes(), 0644) //#nosec
	if err != nil {
		return err
	}

	return nil
}

func getYmlFilePath(platform string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var ymlFilePath string
	if platform == "common" {
		if providedRepoSecretFlags.SecretProd {
			ymlFilePath = cwd + "/" + secretFolder + "/" + platform + "-prod/" + providedRepoSecretFlags.name + ".yml"
		} else {
			ymlFilePath = cwd + "/" + secretFolder + "/" + platform + "-shared/" + providedRepoSecretFlags.name + ".yml"
		}
	} else {
		ymlFilePath = cwd + "/" + secretFolder + "/" + platform + "/" + providedRepoSecretFlags.name + ".yml"
	}

	return ymlFilePath, nil
}

func init() {
	// All environment variables prefixed with below value will be read
	viper.SetEnvPrefix("yak_secret")
	_ = viper.BindEnv("config")

	// Set config file to be read at command execution
	secretCmd.PersistentFlags().StringVarP(&providedRepoSecretFlags.cfgFile, "config", "c", "", "config file")

	secretCmd.PersistentFlags().StringVarP(&providedRepoSecretFlags.Platform, "platform", "P", "", "platform under which the secret is stored")
	secretCmd.PersistentFlags().StringVarP(&providedRepoSecretFlags.Environment, "environment", "e", "", "environment under which the secret is stored")
	secretCmd.PersistentFlags().StringVarP(&providedRepoSecretFlags.Path, "path", "p", "", "path where the secret is stored")
	secretCmd.PersistentFlags().IntVarP(&providedRepoSecretFlags.Version, "version", "v", 0, "secret version")
	secretCmd.PersistentFlags().StringVarP(&providedRepoSecretFlags.team, "team", "t", "", "feature team name managing the secret")

	secretCmd.AddCommand(secretBumpCmd)
	secretCmd.AddCommand(secretCheckCmd)
	secretCmd.AddCommand(secretAddCmd)
	secretCmd.AddCommand(secretUpdateCmd)
	secretCmd.AddCommand(secretFmtCmd)
	secretCmd.AddCommand(secretDeleteCmd)
}

func internalInitConfig(cmd *cobra.Command, args []string) {
	if viper.GetString("config") != "" {
		// Use config file from flag or environment variable.
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		// Search config in various directories with name and extension configured above.
		viper.SetConfigName("secret")
		viper.SetConfigType("yml")

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.yak/")
	}

	viper.AutomaticEnv()

	// Find and read the config file, exit in case of read errors.
	err := viper.ReadInConfig()
	cobra.CheckErr(err)
	log.Debugln("Using config file:", viper.ConfigFileUsed())
}
