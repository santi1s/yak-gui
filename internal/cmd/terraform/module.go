package terraform

import "github.com/spf13/cobra"

var (
	moduleCmd = &cobra.Command{
		Use:   "module",
		Short: "module management",
	}
)

const repositoriesCacheDir = ".cache/yak/repositories"

func init() {
	moduleCmd.AddCommand(moduleBumpCmd)
	moduleCmd.AddCommand(moduleCheckCmd)
	moduleCmd.AddCommand(moduleDependencyCmd)
}
