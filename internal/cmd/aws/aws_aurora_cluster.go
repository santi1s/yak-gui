package aws

import (
	"github.com/spf13/cobra"
)

var (
	auroraClusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "aurora cluster management",
	}
	clusterProvidedFlags clusterFlags
)

type clusterFlags struct {
	targets []string
}

func init() {
	auroraClusterCmd.PersistentFlags().StringArrayVarP(&clusterProvidedFlags.targets, "target", "n", []string{}, "Name of target DB cluster(s).")
	_ = auroraClusterCmd.MarkFlagRequired("target")

	auroraClusterCmd.AddCommand(auroraClusterDeleteCmd)
}
