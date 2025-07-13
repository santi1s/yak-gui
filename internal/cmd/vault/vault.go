package vault

import (
	"os"

	"github.com/doctolib/yak/internal/constant"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type vaultFlags struct {
	admin    bool
	cfgFile  string
	platform string
	cluster  string
	region   string
	json     bool
	team     string
	yaml     bool
}

var (
	providedFlags vaultFlags
	initConfig    = internalInitConfig

	vaultCmd = &cobra.Command{
		Use:   "vault",
		Short: "manage vault infrastructure",
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
	return vaultCmd
}

func init() {
	// All environment variables prefixed with below value will be read
	viper.SetEnvPrefix("yak_vault")
	_ = viper.BindEnv("config")

	// Set config file to be read at command execution
	vaultCmd.PersistentFlags().StringVarP(&providedFlags.cfgFile, "config", "c", "", "config file")
	vaultCmd.PersistentFlags().BoolVar(&providedFlags.json, "json", false, "format output in JSON")
	vaultCmd.PersistentFlags().BoolVar(&providedFlags.yaml, "yaml", false, "format output in YAML")
	_ = viper.BindPFlag("config", vaultCmd.PersistentFlags().Lookup("config"))
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
