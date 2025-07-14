package kafka

import (
	"github.com/santi1s/yak/cli"
	"github.com/spf13/cobra"
)

// formatOutput format the kafka replication health status report to be readable
func formatOutput(comparer string, primaryClusterName string, secondaryClusterName string, inPrimaryButNotInSecondary []string, inSecondaryButNotInPrimary []string) error {
	cli.Printf("\n** %s Replication Status **\n", comparer)
	if len(inPrimaryButNotInSecondary)+len(inSecondaryButNotInPrimary) == 0 {
		cli.Println("No diff found")
		return nil
	}
	const nameWidth = 75
	const primaryPresentWidth = 30
	const secondaryPresentWidth = 15
	cli.Printf("OK: The %s is present on the kafka cluster\n", comparer)
	cli.Printf("NOK: The %s is not present on the kafka cluster\n", comparer)
	cli.Printf("%-*s %-*s %-*s\n", nameWidth, "NAME", primaryPresentWidth, primaryClusterName, secondaryPresentWidth, secondaryClusterName)
	for _, name := range inPrimaryButNotInSecondary {
		cli.Printf("%-*s %-*s %-*s \n", nameWidth, name, primaryPresentWidth, "OK", secondaryPresentWidth, "NOK")
	}
	for _, name := range inSecondaryButNotInPrimary {
		cli.Printf("%-*s %-*s %-*s \n", nameWidth, name, primaryPresentWidth, "NOK", secondaryPresentWidth, "OK")
	}
	return nil
}

func replicationStatus(_ *cobra.Command, _ []string) error {
	primaryClusterName := ProvidedKafkaReplicationFlags.PrimaryCluster
	secondaryClusterName := ProvidedKafkaReplicationFlags.SecondaryCluster
	roleBasedAuth := ProvidedKafkaFlags.roleBasedAuth
	primaryAwsRegion := ProvidedKafkaReplicationFlags.PrimaryAwsRegion
	secondaryAwsRegion := ProvidedKafkaReplicationFlags.SecondaryAwsRegion
	cfgFile := ProvidedKafkaFlags.cfgFile
	topicsInPrimaryButNotInSecondary, topicsInSecondaryButNotInPrimary, err := CompareTopics(primaryClusterName, secondaryClusterName, primaryAwsRegion, secondaryAwsRegion, ProvidedKafkaReplicationFlags.topicsRegexExcluded, cfgFile, roleBasedAuth)
	if err != nil {
		return err
	}
	consumerGroupsInPrimaryButNotInSecondary, consumerGroupsInSecondaryButNotInPrimary, err := CompareConsumerGroups(primaryClusterName, secondaryClusterName, primaryAwsRegion, secondaryAwsRegion, ProvidedKafkaReplicationFlags.consumerGroupsRegexExcluded, cfgFile, roleBasedAuth)
	if err != nil {
		return err
	}
	err = formatOutput("Topic", primaryClusterName, secondaryClusterName, topicsInPrimaryButNotInSecondary, topicsInSecondaryButNotInPrimary)
	if err != nil {
		return err
	}
	err = formatOutput("Consumer Group", primaryClusterName, secondaryClusterName, consumerGroupsInPrimaryButNotInSecondary, consumerGroupsInSecondaryButNotInPrimary)
	if err != nil {
		return err
	}
	return nil
}

var kafkaReplicationHealthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Report about the replication health status of two kafka clusters",
	RunE:  replicationStatus,
	PreRun: func(cmd *cobra.Command, _ []string) {
		err := cmd.Parent().MarkPersistentFlagRequired("primary-cluster")
		if err != nil {
			panic(err)
		}

		err = cmd.Parent().MarkPersistentFlagRequired("secondary-cluster")
		if err != nil {
			panic(err)
		}
		err = cmd.Parent().MarkPersistentFlagRequired("primary-aws-region")
		if err != nil {
			panic(err)
		}
		err = cmd.Parent().MarkPersistentFlagRequired("secondary-aws-region")
		if err != nil {
			panic(err)
		}
	},
	Args: cobra.ExactArgs(0),
}
