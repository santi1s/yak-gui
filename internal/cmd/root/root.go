package root

import (
	"os"

	"github.com/doctolib/yak/internal/cmd/argocd"
	"github.com/doctolib/yak/internal/cmd/aws"
	"github.com/doctolib/yak/internal/cmd/certificate"
	"github.com/doctolib/yak/internal/cmd/couchbase"
	"github.com/doctolib/yak/internal/cmd/github"
	"github.com/doctolib/yak/internal/cmd/helm"
	"github.com/doctolib/yak/internal/cmd/jira"
	"github.com/doctolib/yak/internal/cmd/kafka"
	"github.com/doctolib/yak/internal/cmd/repo"
	"github.com/doctolib/yak/internal/cmd/rollouts"
	"github.com/doctolib/yak/internal/cmd/secret"
	"github.com/doctolib/yak/internal/cmd/terraform"
	"github.com/doctolib/yak/internal/cmd/tfe"
	"github.com/doctolib/yak/internal/cmd/vault"
	"github.com/doctolib/yak/internal/cmd/version"
	"github.com/doctolib/yak/internal/constant"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   constant.CliName,
	Short: constant.CliName + " CLI",
	Long: `CLI for tools maintained by SRE Green team.
More information on https://github.com/doctolib/yak`,
	PersistentPreRun: persistentPreRun,
}

var verboseFlag bool

func persistentPreRun(cmd *cobra.Command, args []string) {
	checkForNewVersion(cmd)
	setLogLevel()
}

func checkForNewVersion(cmd *cobra.Command) {
	if cmd.HasParent() && cmd.Parent().Name() != "completion" {
		version.PrintUpgradeMessage()
	}
}

func setLogLevel() {
	if verboseFlag {
		log.SetLevel(log.DebugLevel)
	}
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	log.SetLevel(logLevel)
}

func init() {
	rootCmd.AddCommand(argocd.GetRootCmd())
	rootCmd.AddCommand(aws.GetRootCmd())
	rootCmd.AddCommand(couchbase.GetRootCmd())
	rootCmd.AddCommand(repo.GetRootCmd())
	rootCmd.AddCommand(rollouts.GetRootCmd())
	rootCmd.AddCommand(secret.GetRootCmd())
	rootCmd.AddCommand(vault.GetRootCmd())
	rootCmd.AddCommand(terraform.GetRootCmd())
	rootCmd.AddCommand(tfe.GetRootCmd())
	rootCmd.AddCommand(version.GetRootCmd())
	rootCmd.AddCommand(certificate.GetRootCmd())
	rootCmd.AddCommand(kafka.GetRootCmd())
	rootCmd.AddCommand(jira.GetRootCmd())
	rootCmd.AddCommand(github.GetRootCmd())
	rootCmd.AddCommand(helm.GetRootCmd())
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "verbose mode")
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
