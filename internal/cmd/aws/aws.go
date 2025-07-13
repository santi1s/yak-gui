package aws

import (
	"github.com/doctolib/yak/internal/constant"
	"github.com/spf13/cobra"
)

type awsFlags struct {
	cfgFile     string
	description string
	check       bool
	skipConfirm bool
	dryRun      bool
}

var (
	awsCmd = &cobra.Command{
		Use:   "aws",
		Short: "manage aws",
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
	return awsCmd
}

func init() {
	awsCmd.AddCommand(auroraCmd)
	awsCmd.AddCommand(configCmd)
}
