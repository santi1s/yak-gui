package tfe

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/doctolib/yak/internal/constant"
	gotfe "github.com/hashicorp/go-tfe"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type tfeFlags struct {
	cfgFile string
}

var (
	initConfig    = internalInitConfig
	providedFlags tfeFlags
	tfeCmd        = &cobra.Command{
		Use:   "tfe",
		Short: "manage tfe",
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

			initConfig(cmd, args)
			cmd.SilenceUsage = true
		},
	}
)

func GetRootCmd() *cobra.Command {
	return tfeCmd
}

func init() {
	// All environment variables prefixed with below value will be read
	viper.SetEnvPrefix("yak_tfe")
	_ = viper.BindEnv("config")

	// Set config file to be read at command execution
	tfeCmd.PersistentFlags().StringVarP(&providedFlags.cfgFile, "config", "c", "", "config file")
	_ = viper.BindPFlag("config", tfeCmd.PersistentFlags().Lookup("config"))

	tfeCmd.AddCommand(tfePlanCmd)
	tfeCmd.AddCommand(tfeVersionCmd)
	tfeCmd.AddCommand(tfeCheckVersionCmd)
	tfeCmd.AddCommand(tfeWorkspaceCmd)
	tfeCmd.AddCommand(tfeRunCmd)
}

func internalInitConfig(cmd *cobra.Command, args []string) {
	if viper.GetString("config") != "" {
		// Use config file from flag or environment variable.
		viper.SetConfigFile(viper.GetString("config"))
	} else {
		// Search config in various directories with name and extension configured above.
		viper.SetConfigName("tfe")
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

// Errors
var errTfeTokenNotFound = errors.New("you must have a TFE_TOKEN environment variable set or a $HOME/.terraform.d/credentials.tfrc.json with your TFE API token")

type TerraformConfig struct {
	Token string `json:"token"`
}

type TerraformCredentials struct {
	Credentials map[string]TerraformConfig `json:"credentials"`
}

func getTfeConfig() (*gotfe.Config, error) {
	endpoint := viper.GetString("endpoint")

	tfeToken := os.Getenv("TF_TOKEN_" + strings.ReplaceAll(strings.ReplaceAll(endpoint, ".", "_"), "-", "_"))
	if tfeToken == "" {
		tfeToken = os.Getenv("TFE_TOKEN")
		if tfeToken == "" {
			home := os.Getenv("HOME")
			if home != "" {
				credentialsJSON, err := os.ReadFile(path.Join(home, ".terraform.d", "credentials.tfrc.json"))
				if err != nil {
					return nil, errTfeTokenNotFound
				}

				credentials := TerraformCredentials{}
				err = json.Unmarshal(credentialsJSON, &credentials)
				if err != nil {
					return nil, errTfeTokenNotFound
				}

				if val, ok := credentials.Credentials[endpoint]; ok {
					tfeToken = val.Token
				} else {
					return nil, errTfeTokenNotFound
				}
			} else {
				return nil, errTfeTokenNotFound
			}
		}
	}

	return &gotfe.Config{
		Address: fmt.Sprintf("https://%s", endpoint),
		Token:   tfeToken,
	}, nil
}

func getTfeClient(config *gotfe.Config) (*gotfe.Client, error) {
	return gotfe.NewClient(config)
}
