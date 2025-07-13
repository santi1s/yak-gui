package secret

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/constant"
	"github.com/doctolib/yak/internal/helper"
	"github.com/hashicorp/vault/api"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type secretFlags struct {
	cfgFile       string
	platform      string
	environment   string
	path          string
	version       int
	json          bool
	yaml          bool
	keys          []string
	key           string
	value         string
	skipConfirm   bool
	owner         string
	team          string
	source        string
	usage         string
	remove        bool
	interactive   bool
	baseVersion   int
	diffVersion   int
	allPlatforms  bool
	localName     string
	targetService string
	secret        string
	serviceName   string
	clientName    string
	clientSecret  string
	dataKey       string
	force         bool
}

type Output struct {
	Stdout io.Writer
	Stderr io.Writer
}

var (
	initConfig    = internalInitConfig
	providedFlags secretFlags
	secretCmd     = &cobra.Command{
		Use:   "secret",
		Short: "manage a secret in vault",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
				cmd.Root().PersistentPreRun(cmd, args)
			}

			if cmd.HasParent() && cmd.Parent().Name() == "completion" {
				switch cmd.Name() {
				case "bash", "zsh", "fish", "powershell":
					cmd.ResetFlags()
				}
			}

			if cmd.HasParent() && cmd.Parent().Name() == "secret" {
				_ = viper.BindPFlag("config", cmd.Parent().PersistentFlags().Lookup("config"))
			}

			if cmd.HasParent() && cmd.Parent().HasParent() && cmd.Parent().Parent().Name() == "secret" {
				_ = viper.BindPFlag("config", cmd.Parent().Parent().PersistentFlags().Lookup("config"))
			}

			initConfig(cmd, args)
			cmd.SilenceUsage = true
		},
	}

	// Parameter errors
	errOwnerCantBeEmpty                 = errors.New("owner flag can't be empty")
	errOwnerMismatch                    = errors.New("owner mismatch while updating secret")
	errSourceCantBeEmpty                = errors.New("source flag can't be empty")
	errUsageCantBeEmpty                 = errors.New("usage flag can't be empty")
	errKeyParameterCantBeEmpty          = errors.New("key parameter can't be empty")
	errValueParameterCantBeEmpty        = errors.New("value parameter can't be empty")
	errKeysCantBeSetWithoutRemove       = errors.New("keys flag can't be set without remove flag being set")
	errEmptyInput                       = errors.New("input can't be empty")
	errCantContainEmptyValue            = errors.New("input cannot contain any empty value")
	errCantContainValue                 = errors.New("input cannot contain any value")
	errAskConfirmationNotConfirmed      = errors.New("action not confirmed by user")
	errInteractiveKeysCantBeSetTogether = errors.New("interactive flag and keys cannot be set together")
	errMetadataNotDeletable             = errors.New("this metadata is mandatory and can't be deleted")
	errMetadataCouldNotBeAdded          = errors.New("secret is created but metadata could not be added")
	errSameVersionDiff                  = errors.New("no need to compare the same versions of a secret")

	// Read operation errors
	errSecretAlreadyExists   = errors.New("secret already exists")
	errMetadataKeyNotFound   = errors.New("metadata key not found, create the metadata before update it")
	errMetadataEmpty         = errors.New("secret's metadata are empty, this should be impossible")
	errMetadataAlreadyExists = errors.New("secret's metadata already exists")

	// Check errors
	errSecretsNotSynced      = errors.New("some secrets are not synced with their ci counterpart")
	errCheckSyncMissingFlags = errors.New("you must provide --all-platforms or --platform flag")
	errInvalidJWTSecret      = errors.New("secret is invalid for HMAC-SHA256")
	errInvalidJSONSecretData = errors.New("invalid JSON in existing secret data")
)

func GetRootCmd() *cobra.Command {
	return secretCmd
}

func init() {
	// All environment variables prefixed with below value will be read
	viper.SetEnvPrefix("yak_secret")
	_ = viper.BindEnv("config")

	// Set config file to be read at command execution
	secretCmd.PersistentFlags().StringVarP(&providedFlags.cfgFile, "config", "c", "", "config file")

	// Set available flags
	secretCmd.PersistentFlags().StringVarP(&providedFlags.platform, "platform", "P", "", "platform under which the secret is stored")
	secretCmd.PersistentFlags().StringVarP(&providedFlags.environment, "environment", "e", "", "environment under which the secret is stored")
	secretCmd.PersistentFlags().StringVarP(&providedFlags.path, "path", "p", "", "path where the secret is stored (mandatory)")
	secretCmd.PersistentFlags().StringVarP(&providedFlags.team, "team", "t", "", "the name of the feature team managing the secret")
	secretCmd.PersistentFlags().BoolVar(&providedFlags.json, "json", false, "format output in JSON")
	secretCmd.PersistentFlags().BoolVar(&providedFlags.yaml, "yaml", false, "format output in YAML")
	secretCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	err := secretCmd.MarkPersistentFlagRequired("path")
	if err != nil {
		panic(err)
	}

	secretCmd.AddCommand(metadataCmd)
	secretCmd.AddCommand(createCmd)
	secretCmd.AddCommand(deleteCmd)
	secretCmd.AddCommand(getCmd)
	secretCmd.AddCommand(updateCmd)
	secretCmd.AddCommand(diffCmd)
	secretCmd.AddCommand(checkSyncCmd)
	secretCmd.AddCommand(resyncCmd)
	secretCmd.AddCommand(listCmd)
	secretCmd.AddCommand(listDuplicatesCmd)
	secretCmd.AddCommand(jwtCmd)
	secretCmd.AddCommand(undeleteCmd)
	secretCmd.AddCommand(cleanVaultCmd)
	secretCmd.AddCommand(destroyCmd)
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

// Returns latest non deleted version of a secret
// -1 with no error means that the secret does not exist or that all versions of the secret are deleted
// so there is no latest non deleted version
// -1 with an error means there have been an unrecoverable error
func GetLatestVersion(clients []*api.Client, secretPath string) (int, error) {
	secret, err := ReadSecretMetadata(clients, secretPath)
	if err != nil {
		if err == helper.ErrSecretNotFound {
			return -1, nil
		}
		return -1, err
	}

	currentVersion, err := strconv.Atoi(secret.Data["current_version"].(json.Number).String())
	if err != nil {
		return -1, err
	}

	oldestVersion, err := strconv.Atoi(secret.Data["oldest_version"].(json.Number).String())
	if err != nil {
		return -1, err
	}

	versions := secret.Data["versions"].(map[string]interface{})
	version := 0
	for i := currentVersion; i > oldestVersion; i-- {
		if versions[strconv.Itoa(i)].(map[string]interface{})["deletion_time"] == "" && versions[strconv.Itoa(i)].(map[string]interface{})["destroyed"] == false {
			version = i
			break
		}
	}
	if version == 0 {
		return -1, nil
	}

	return version, nil
}

func formatOutput(output map[string]interface{}) error {
	switch {
	case providedFlags.json:
		return cli.PrintJSON(output)
	case providedFlags.yaml:
		return cli.PrintYAML(output)
	default:
		return cli.PrintYAML(output)
	}
}

func readFromStdin() (map[string]interface{}, error) {
	kvs := make(map[string]string)
	input, err := cli.ReadAll()
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(input), &kvs)
	if err != nil {
		return nil, err
	}

	if len(kvs) == 0 {
		return nil, errEmptyInput
	}

	kvi := make(map[string]interface{}, len(kvs))
	for k, v := range kvs {
		kvi[k] = v
	}

	return kvi, nil
}

func readFromInteractive(keysOnly bool) (map[string]interface{}, error) {
	var key string
	var value string
	var confirmationPrompt string

	data := make(map[string]interface{})

	for {
		value = ""
		// Read key
		for {
			key = ""
			cli.Print("Enter a key: ")
			rawKey, err := cli.Read()
			if err != nil && err.Error() != "unexpected newline" {
				_, _ = cli.PrintlnErr(err)
				continue
			}
			key = strings.TrimSpace(string(rawKey))
			if _, ok := data[key]; ok {
				cli.Println("Key already inserted, please use another one")
			} else if key != "" {
				break
			}
		}
		if !keysOnly {
			// Read value
			for {
				cli.Print("Enter a value: ")
				rawValue, err := cli.ReadPassword()
				if err != nil && err.Error() != "unexpected newline" {
					_, _ = cli.PrintlnErr(err)
					continue
				}
				value = strings.TrimSpace(string(rawValue))
				cli.Println()
				if value != "" {
					break
				}
			}
			confirmationPrompt = "Do you want to add another key/pair value?"
		} else {
			confirmationPrompt = "Do you want to add another key?"
		}
		// Create entry
		data[key] = value
		// Ask for more key-value pairs
		if !cli.AskConfirmation(confirmationPrompt) {
			break
		}
	}
	return data, nil
}

func hasEmptyValue(m map[string]interface{}) bool {
	for _, v := range m {
		if v == "" || v == nil {
			return true
		}
	}
	return false
}

func hasValue(m map[string]interface{}) bool {
	for _, v := range m {
		if v != "" {
			return true
		}
	}
	return false
}
