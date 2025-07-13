package github

import (
	"github.com/doctolib/yak/internal/constant"
	"github.com/spf13/cobra"
)

// type ghFlags struct {
// 	jiraURL string
// }

var (
	// ghFlags jiraFlags

	ghCmd = &cobra.Command{
		Use:   "github",
		Short: "Github related commands",
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
	return ghCmd
}

func init() {
	ghCmd.AddCommand(prCmd)
}
