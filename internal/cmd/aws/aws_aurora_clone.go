package aws

import (
	"github.com/spf13/cobra"
)

var (
	auroraCloneCmd = &cobra.Command{
		Use:              "clone",
		Short:            "aurora clone management",
		PersistentPreRun: sortCloneFlags,
	}
	cloneProvidedFlags cloneFlags
)

type cloneFlags struct {
	pr, size                string
	sources                 []string
	targets                 []string
	tags                    map[string]string
	fromSnapshot            bool
	subnetGroupName         string
	dbClusterParameterGroup []string
}

func init() {
	auroraCloneCmd.PersistentFlags().StringVarP(&cloneProvidedFlags.pr, "pr", "p", "", "Pull Request number")
	auroraCloneCmd.PersistentFlags().StringArrayVarP(&cloneProvidedFlags.sources, "source", "s", []string{}, "DB source(s) to clone from. Number of sources must match number of target names.")
	auroraCloneCmd.PersistentFlags().StringArrayVarP(&cloneProvidedFlags.targets, "target", "n", []string{}, "Name of target DB source(s). Number of target names must match number of source(s)")
	auroraCloneCmd.PersistentFlags().StringVarP(&cloneProvidedFlags.size, "size", "z", "db.r5.large", "Clone instance size")
	auroraCloneCmd.PersistentFlags().StringToStringVarP(&cloneProvidedFlags.tags, "tags", "t", map[string]string{}, "Tags to apply to the clone")
	auroraCloneCmd.PersistentFlags().BoolVar(&cloneProvidedFlags.fromSnapshot, "from-snapshot", false, "Create the clone from a snapshot instead of a point in time")
	auroraCloneCmd.PersistentFlags().StringVar(&cloneProvidedFlags.subnetGroupName, "subnet-group-name", "default", "Target subnet group name. Only used when using --from-snapshot")
	auroraCloneCmd.PersistentFlags().StringArrayVar(&cloneProvidedFlags.dbClusterParameterGroup, "dbcluster-parameter-group", []string{}, "Target DBClusterParameterGroup for each instance to create. Only used when using --from-snapshot. Number of dbClusterParameterGroup must match number of source(s).")

	_ = auroraCloneCmd.MarkFlagRequired("source")
	auroraCloneCmd.MarkFlagsMutuallyExclusive("from-snapshot", "pr")

	auroraCloneCmd.AddCommand(auroraCloneCreateCmd)
	auroraCloneCmd.AddCommand(auroraCloneDeleteCmd)
}
