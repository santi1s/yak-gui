package rollouts

import (
	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
)

type rolloutsFlags struct {
	addr      string
	namespace string
	json      bool
	yaml      bool
}

var (
	providedFlags rolloutsFlags

	rolloutsCmd = &cobra.Command{
		Use:   "rollouts",
		Short: "A suite of commands to manage Argo Rollouts resources",
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
	return rolloutsCmd
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

// resolveNamespace returns the effective namespace to use:
// 1. If explicitly provided via --namespace flag, use that
// 2. Otherwise, try to get namespace from current kubeconfig context
// 3. Fall back to "default" if neither is available
func resolveNamespace() string {
	if providedFlags.namespace != "" {
		return providedFlags.namespace
	}

	// Try to get namespace from kubeconfig context
	if contextNamespace, err := helper.GetKubernetesCurrentNamespace(); err == nil && contextNamespace != "" {
		return contextNamespace
	}

	// Fall back to default
	return "default"
}

func init() {
	rolloutsCmd.PersistentFlags().StringVarP(&providedFlags.addr, "server", "s", "", "Kubernetes API server address")
	rolloutsCmd.PersistentFlags().StringVarP(&providedFlags.namespace, "namespace", "n", "", "Kubernetes namespace (defaults to current kubeconfig context namespace)")
	rolloutsCmd.PersistentFlags().BoolVar(&providedFlags.json, "json", false, "format output in JSON")
	rolloutsCmd.PersistentFlags().BoolVar(&providedFlags.yaml, "yaml", false, "format output in YAML")
	rolloutsCmd.AddCommand(statusCmd)
	rolloutsCmd.AddCommand(getCmd)
	rolloutsCmd.AddCommand(listCmd)
	rolloutsCmd.AddCommand(promoteCmd)
	rolloutsCmd.AddCommand(pauseCmd)
	rolloutsCmd.AddCommand(abortCmd)
	rolloutsCmd.AddCommand(restartCmd)
	rolloutsCmd.AddCommand(analysisCmd)
	rolloutsCmd.AddCommand(logsCmd)
	rolloutsCmd.AddCommand(historyCmd)
	rolloutsCmd.AddCommand(retryCmd)
	rolloutsCmd.AddCommand(setImageCmd)
	rolloutsCmd.AddCommand(undoCmd)
}
