package argocd

import (
	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	"github.com/spf13/cobra"
)

type argocdFlags struct {
	addr    string
	project string
	json    bool
	yaml    bool
}

var (
	providedFlags argocdFlags

	argocdCmd = &cobra.Command{
		Use:   "argocd",
		Short: "A suite of commands to manage argocd related resources",
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

			cmd.SilenceUsage = true
		},
	}
)

func GetRootCmd() *cobra.Command {
	return argocdCmd
}

func formatOutput(output interface{}) error {
	switch {
	case providedFlags.json:
		return cli.PrintJSON(output)
	case providedFlags.yaml:
		return cli.PrintYAML(output)
	default:
		return cli.PrintYAML(output)
	}
}

func init() {
	argocdCmd.PersistentFlags().StringVarP(&providedFlags.addr, "argocd-addr", "s", "", "Argocd server")
	argocdCmd.PersistentFlags().BoolVar(&providedFlags.json, "json", false, "format output in JSON")
	argocdCmd.PersistentFlags().BoolVar(&providedFlags.yaml, "yaml", false, "format output in YAML")
	argocdCmd.PersistentFlags().StringVarP(&providedFlags.project, "project", "j", "main", "Project name")
	argocdCmd.AddCommand(monitorArgoCDCmd)
	argocdCmd.AddCommand(suspendCmd)
	argocdCmd.AddCommand(unsuspendCmd)
	argocdCmd.AddCommand(statusCmd)
	argocdCmd.AddCommand(orphanResourcesCmd)
	argocdCmd.AddCommand(syncCmd)
	argocdCmd.AddCommand(refreshCmd)
	argocdCmd.AddCommand(diffCmd)
	argocdCmd.AddCommand(pruneCmd)
	argocdCmd.AddCommand(resourcesCmd)
	argocdCmd.AddCommand(logsCmd)
	argocdCmd.AddCommand(getCmd)
	argocdCmd.AddCommand(dashboardCmd)
}
