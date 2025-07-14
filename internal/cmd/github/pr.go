package github

import (
	"github.com/santi1s/yak/internal/constant"
	"github.com/spf13/cobra"
)

var (
	prCmd = &cobra.Command{
		Use:   "pr",
		Short: "Pull Request related commands",
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

func init() {
	prCmd.AddCommand(createPullRequestCmd)
}
