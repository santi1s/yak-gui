package terraform

import "github.com/spf13/cobra"

var (
	providerCmd = &cobra.Command{
		Use:   "provider",
		Short: "provider management",
	}
)

type providerFlags struct {
}

func init() {
	providerCmd.AddCommand(providerCheckCmd)
	providerCmd.AddCommand(providerBumpCmd)
	providerCmd.AddCommand(providerGenerateCmd)
}
