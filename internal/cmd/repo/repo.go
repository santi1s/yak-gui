package repo

import (
	"github.com/doctolib/yak/internal/constant"
	"github.com/spf13/cobra"
)

type kubeFlags struct {
}

var (
	providedFlags kubeFlags

	repoCmd = &cobra.Command{
		Use:     "repo",
		Aliases: []string{"repository", "kube", "terraform-infra"},
		Short:   "tools to manage kube and terraform-infra repository",
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
	return repoCmd
}

func init() {
	repoCmd.AddCommand(secretCmd)
}
