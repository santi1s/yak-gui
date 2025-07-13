package certificate

import (
	"errors"
	"os"
	"strings"

	"github.com/doctolib/yak/internal/constant"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type certificateFlags struct {
	cfgFile              string
	certificate          string
	dryRun               bool
	jiraTicket           string
	dcvOnly              bool
	version, diffVersion int
}

var (
	initConfig     = internalInitConfig
	providedFlags  certificateFlags
	certificateCmd = &cobra.Command{
		Use:   "certificate",
		Short: "manage certificate",
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
	errCertificateCantBeEmpty = errors.New("--certificate flag can't be empty")
	errJiraTicketCantBeEmpty  = errors.New("--jira flag can't be empty")
)

func GetRootCmd() *cobra.Command {
	return certificateCmd
}

func init() {
	// All environment variables prefixed with below value will be read
	viper.SetEnvPrefix("yak_certificate")
	_ = viper.BindEnv("config")

	// Set config file to be read at command execution
	certificateCmd.PersistentFlags().StringVarP(&providedFlags.cfgFile, "config", "c", "", "config file")
	_ = viper.BindPFlag("config", certificateCmd.PersistentFlags().Lookup("config"))
	certificateCmd.PersistentFlags().StringVarP(&providedFlags.certificate, "certificate", "C", "", "certificate")
	_ = viper.BindPFlag("certificate", certificateCmd.PersistentFlags().Lookup("certificate"))
	certificateCmd.PersistentFlags().StringVarP(&providedFlags.certificate, "jira", "j", "", "Jira ticket")
	_ = viper.BindPFlag("jira", certificateCmd.PersistentFlags().Lookup("jira"))

	certificateCmd.AddCommand(certificateRenewCmd)
	certificateCmd.AddCommand(certificateRefreshSecretCmd)
	certificateCmd.AddCommand(certificateDescribeSecretCmd)
	certificateCmd.AddCommand(certificateGandiCheckCmd)
}

type CloudflareConfig struct {
	Path string `yaml:"path"`
	Zone string `yaml:"zone"`
}

type Route53Config struct {
	Path string `yaml:"path"`
	Zone string `yaml:"zone"`
}

type SecretConfig struct {
	Platform string      `yaml:"platform"`
	Env      string      `yaml:"env"`
	Path     string      `yaml:"path"`
	Keys     *KeysConfig `yaml:"keys"`
}

type KeysConfig struct {
	Certificate  string `yaml:"certificate"`
	PrivateKey   string `yaml:"private_key"`
	Intermediate string `yaml:"intermediate"`
}

type CertConfig struct {
	Name       string            `yaml:"name"`
	Conf       string            `yaml:"conf"`
	Issuer     string            `yaml:"issuer"`
	Tags       []string          `yaml:"tags"`
	Cloudflare *CloudflareConfig `yaml:"cloudflare"`
	Route53    *Route53Config    `yaml:"route53"`
	Secret     *SecretConfig     `yaml:"secret"`
}

func (c CertConfig) GetTfFilePath() (string, error) {
	if c.Cloudflare != nil {
		return c.Cloudflare.Path, nil
	} else if c.Route53 != nil {
		return c.Route53.Path, nil
	}
	return "", errors.New("no path found for certificate " + c.Name)
}

func (c CertConfig) DNSProvider() (string, error) {
	if c.Cloudflare != nil {
		return "cloudflare", nil
	} else if c.Route53 != nil {
		return "route53", nil
	}
	return "", errors.New("no DNS provider found for certificate " + c.Name)
}

func readCertificatesConfig() ([]CertConfig, error) {
	ymlPath := os.Getenv("TFINFRA_REPOSITORY_PATH") + "/sslcerts/config.yml"

	file, err := os.ReadFile(ymlPath)
	if err != nil {
		return nil, err
	}

	var certificatesConfig []CertConfig
	err = yaml.Unmarshal(file, &certificatesConfig)

	if err != nil {
		return nil, err
	}

	return certificatesConfig, nil
}

func internalInitConfig(_ *cobra.Command, _ []string) {
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

	err := viper.ReadInConfig()
	cobra.CheckErr(err)
	log.Debugln("Using config file:", viper.ConfigFileUsed())
}

func getCertificateConfig() (*CertConfig, error) {
	certificateConfigs, err := readCertificatesConfig()
	if err != nil {
		return nil, err
	}

	for _, config := range certificateConfigs {
		name := config.Name
		if name == providedFlags.certificate {
			certConfig := config

			// Exit if there is no issuer, we don't know what API to use
			if certConfig.Issuer == "" {
				return nil, errors.New("the configuration should have the key 'issuer'")
			}
			// Exit if there is no tags, we don't know which resource to use in API
			if len(certConfig.Tags) == 0 {
				return nil, errors.New("the configuration should have the key 'tags'")
			}
			// Exit if there is no secret, we don't know what to update
			if certConfig.Secret == nil {
				return nil, errors.New("the configuration should have the key 'secret'")
			}
			if certConfig.Secret.Path == "" {
				return nil, errors.New("the configuration should have the key 'secret.path'")
			}
			if certConfig.Secret.Platform == "" {
				return nil, errors.New("the configuration should have the key 'secret.platform'")
			}
			// No certificate key name, we don't know in which key to put the certificate
			if certConfig.Secret.Keys == nil {
				return nil, errors.New("the configuration should have the key 'secret.keys'")
			}
			if certConfig.Secret.Keys.Certificate == "" {
				return nil, errors.New("the configuration should have the key 'secret.keys.certificate'")
			}
			// No private key name, we don't know in which key to put the private key
			if certConfig.Secret.Keys.PrivateKey == "" {
				return nil, errors.New("the configuration should have the key 'secret.keys.privatekey'")
			}
			if certConfig.Cloudflare == nil && certConfig.Route53 == nil {
				return nil, errors.New("the configuration should have the key 'cloudflare' or 'route53'")
			}
			if certConfig.Cloudflare != nil && certConfig.Cloudflare.Zone == "" {
				return nil, errors.New("the configuration should have the key 'cloudflare.zone'")
			}
			if certConfig.Cloudflare != nil && certConfig.Cloudflare.Path == "" {
				return nil, errors.New("the configuration should have the key 'cloudflare.path'")
			}
			if certConfig.Cloudflare != nil && !strings.HasSuffix(certConfig.Cloudflare.Path, ".tf") {
				return nil, errors.New("'cloudflare.path' parameter should lead to a .tf file")
			}
			if certConfig.Route53 != nil && certConfig.Route53.Zone == "" {
				return nil, errors.New("the configuration should have the key 'route53.zone'")
			}
			if certConfig.Route53 != nil && certConfig.Route53.Path == "" {
				return nil, errors.New("the configuration should have the key 'route53.path'")
			}
			if certConfig.Route53 != nil && !strings.HasSuffix(certConfig.Route53.Path, ".tf") {
				return nil, errors.New("'route53.path' parameter should lead to a .tf file")
			}
			// Exit if there is no conf, we can't generate the CSR
			if certConfig.Conf == "" {
				return nil, errors.New("the configuration should have the key 'conf'")
			}
			if !strings.HasSuffix(certConfig.Conf, ".conf") {
				return nil, errors.New("conf' parameter should lead to a .conf file")
			}

			return &certConfig, nil
		}
	}
	return nil, errors.New("certificate " + providedFlags.certificate + " not found in config file")
}
