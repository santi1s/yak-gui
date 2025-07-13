package terraform

import "github.com/spf13/cobra"

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Terraform version management",
	}
)

func init() {
	versionCmd.AddCommand(versionCheckCmd)
}
