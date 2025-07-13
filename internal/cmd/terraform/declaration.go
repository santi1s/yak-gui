package terraform

import "github.com/spf13/cobra"

var (
	declarationCmd = &cobra.Command{
		Use:   "declaration",
		Short: "Declarations management", //TODO: find better name
	}
)

func init() {
	declarationCmd.AddCommand(declarationCheckCmd)
}
