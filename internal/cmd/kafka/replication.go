package kafka

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/doctolib/yak/internal/constant"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
)

type KafkaReplicationFlags struct {
	Platform                    string
	PrimaryCluster              string
	PrimaryAwsRegion            string
	SecondaryCluster            string
	SecondaryAwsRegion          string
	topicsRegexExcluded         []string
	consumerGroupsRegexExcluded []string
}

var ProvidedKafkaReplicationFlags KafkaReplicationFlags

var kafkaReplicationCmd = &cobra.Command{
	Use:   "replication",
	Short: "manage replication between two kafka clusters",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
			cmd.Root().PersistentPreRun(cmd, args)
		}
	},
}

// CompareTopics returns the diff in terms of topics present inside the secondary and the primary kafka clusters
// Some topics can be excluded from comparaison through topicsRegexExcluded
// topicsRegexExcluded is an array of regex
func CompareTopics(primaryClusterName string, secondaryClusterName string, PrimaryAwsRegion string, SecondaryAwsRegion string, topicsRegexExcluded []string, cfgFile string, roleBasedAuth bool) ([]string, []string, error) {
	var topicsInPrimaryButNotInSecondary []string
	var topicsInSecondaryButNotInPrimary []string
	var topicsInPrimaryButNotInSecondaryAfterCleanup []string
	var topicsInSecondaryButNotInPrimaryAfterCleanup []string

	primaryTopics, err := helper.GetTopics(primaryClusterName, PrimaryAwsRegion, cfgFile, roleBasedAuth)
	if err != nil {
		return topicsInPrimaryButNotInSecondary, topicsInSecondaryButNotInPrimary, err
	}
	secondaryTopics, err := helper.GetTopics(secondaryClusterName, SecondaryAwsRegion, cfgFile, roleBasedAuth)
	if err != nil {
		return topicsInPrimaryButNotInSecondary, topicsInSecondaryButNotInPrimary, err
	}

	for topicName, topicInfo := range primaryTopics {
		topicInfoSecondary, found := secondaryTopics[topicName]
		if found {
			if (topicInfoSecondary.NumPartitions != topicInfo.NumPartitions) ||
				(topicInfoSecondary.ReplicationFactor != topicInfo.ReplicationFactor) {
				fmt.Println(topicName, " : Topics do not have the same configuration")
			}
		} else {
			topicsInPrimaryButNotInSecondary = append(topicsInPrimaryButNotInSecondary, topicName)
		}
	}
	for topicName, topicInfo := range secondaryTopics {
		topicInfoPrimary, found := primaryTopics[topicName]
		if found {
			if (topicInfoPrimary.NumPartitions != topicInfo.NumPartitions) ||
				(topicInfoPrimary.ReplicationFactor != topicInfo.ReplicationFactor) {
				fmt.Println(topicName, " : Topics do not have the same configuration")
			}
		} else {
			topicsInSecondaryButNotInPrimary = append(topicsInSecondaryButNotInPrimary, topicName)
		}
	}

	for _, topic := range topicsInPrimaryButNotInSecondary {
		topicsInPrimaryButNotInSecondaryAfterCleanup = append(topicsInPrimaryButNotInSecondaryAfterCleanup, topic)
		for _, regexTopic := range topicsRegexExcluded {
			match, _ := regexp.MatchString(regexTopic, topic)
			if match {
				topicsInPrimaryButNotInSecondaryAfterCleanup = topicsInPrimaryButNotInSecondaryAfterCleanup[:len(topicsInPrimaryButNotInSecondaryAfterCleanup)-1]
				break
			}
		}
	}

	for _, topic := range topicsInSecondaryButNotInPrimary {
		topicsInSecondaryButNotInPrimaryAfterCleanup = append(topicsInSecondaryButNotInPrimaryAfterCleanup, topic)
		for _, regexTopic := range topicsRegexExcluded {
			match, _ := regexp.MatchString(regexTopic, topic)
			if match {
				topicsInSecondaryButNotInPrimaryAfterCleanup = topicsInSecondaryButNotInPrimaryAfterCleanup[:len(topicsInSecondaryButNotInPrimaryAfterCleanup)-1]
				break
			}
		}
	}
	sort.Strings(topicsInPrimaryButNotInSecondaryAfterCleanup)
	sort.Strings(topicsInSecondaryButNotInPrimaryAfterCleanup)
	return topicsInPrimaryButNotInSecondaryAfterCleanup, topicsInSecondaryButNotInPrimaryAfterCleanup, nil
}

// CompareConsumerGroups returns the diff in terms of consumer groups present inside the secondary and the primary kafka clusters
// Some consumer groups can be excluded from comparaison through consumerGroupsRegexExcluded
// consumerGroupsRegexExcluded is an array of regex
func CompareConsumerGroups(primaryClusterName string, secondaryClusterName string, PrimaryAwsRegion string, SecondaryAwsRegion string, consumerGroupsRegexExcluded []string, cfgFile string, roleBasedAuth bool) ([]string, []string, error) {
	var inPrimaryButNotInSecondary []string
	var inSecondaryButNotInPrimary []string
	var inPrimaryButNotInSecondaryAfterCleanup []string
	var inSecondaryButNotInPrimaryAfterCleanup []string
	primaryConsumerGroups, err := helper.GetConsumerGroups(primaryClusterName, PrimaryAwsRegion, cfgFile, []string{"consumer"}, roleBasedAuth)
	if err != nil {
		return inPrimaryButNotInSecondary, inSecondaryButNotInPrimary, err
	}
	secondaryConsumerGroups, err := helper.GetConsumerGroups(secondaryClusterName, SecondaryAwsRegion, cfgFile, []string{"consumer"}, roleBasedAuth)
	if err != nil {
		return inPrimaryButNotInSecondary, inSecondaryButNotInPrimary, err
	}
	sort.Strings(primaryConsumerGroups)
	sort.Strings(secondaryConsumerGroups)
	for _, primaryConsumerGroup := range primaryConsumerGroups {
		inPrimaryButNotInSecondary = append(inPrimaryButNotInSecondary, primaryConsumerGroup)
		for _, secondaryConsumerGroup := range secondaryConsumerGroups {
			if primaryConsumerGroup == secondaryConsumerGroup {
				inPrimaryButNotInSecondary = inPrimaryButNotInSecondary[:len(inPrimaryButNotInSecondary)-1]
				break
			}
		}
	}

	for _, secondaryConsumerGroup := range secondaryConsumerGroups {
		inSecondaryButNotInPrimary = append(inSecondaryButNotInPrimary, secondaryConsumerGroup)
		for _, primaryConsumerGroup := range primaryConsumerGroups {
			if primaryConsumerGroup == secondaryConsumerGroup {
				inSecondaryButNotInPrimary = inSecondaryButNotInPrimary[:len(inSecondaryButNotInPrimary)-1]
				break
			}
		}
	}

	for _, consumerGroup := range inPrimaryButNotInSecondary {
		inPrimaryButNotInSecondaryAfterCleanup = append(inPrimaryButNotInSecondaryAfterCleanup, consumerGroup)
		for _, regexConsumerGroup := range consumerGroupsRegexExcluded {
			match, _ := regexp.MatchString(regexConsumerGroup, consumerGroup)
			if match {
				inPrimaryButNotInSecondaryAfterCleanup = inPrimaryButNotInSecondaryAfterCleanup[:len(inPrimaryButNotInSecondaryAfterCleanup)-1]
				break
			}
		}
	}

	for _, consumerGroup := range inSecondaryButNotInPrimary {
		inSecondaryButNotInPrimaryAfterCleanup = append(inSecondaryButNotInPrimaryAfterCleanup, consumerGroup)
		for _, regexConsumerGroup := range consumerGroupsRegexExcluded {
			match, _ := regexp.MatchString(regexConsumerGroup, consumerGroup)
			if match {
				inSecondaryButNotInPrimaryAfterCleanup = inSecondaryButNotInPrimaryAfterCleanup[:len(inSecondaryButNotInPrimaryAfterCleanup)-1]
				break
			}
		}
	}

	return inPrimaryButNotInSecondaryAfterCleanup, inSecondaryButNotInPrimaryAfterCleanup, nil
}

func init() {
	kafkaReplicationCmd.PersistentFlags().StringVarP(&ProvidedKafkaReplicationFlags.PrimaryCluster, "primary-cluster", "p", "", "kafka primary cluster name, example: staging-aws-fr-par-1")
	kafkaReplicationCmd.PersistentFlags().StringVarP(&ProvidedKafkaReplicationFlags.PrimaryAwsRegion, "primary-aws-region", "P", "", "primary aws region name, example: eu-west-3")
	kafkaReplicationCmd.PersistentFlags().StringVarP(&ProvidedKafkaReplicationFlags.SecondaryCluster, "secondary-cluster", "s", "", "kafka secondary cluster name, example: staging-aws-de-fra-1")
	kafkaReplicationCmd.PersistentFlags().StringVarP(&ProvidedKafkaReplicationFlags.SecondaryAwsRegion, "secondary-aws-region", "S", "", "secondary aws region name, example: eu-central-1")
	kafkaReplicationCmd.PersistentFlags().StringArrayVarP(&ProvidedKafkaReplicationFlags.topicsRegexExcluded, "topicsExcluded", "t", []string{}, "excluded topic regexes from replication status")
	kafkaReplicationCmd.PersistentFlags().StringArrayVarP(&ProvidedKafkaReplicationFlags.consumerGroupsRegexExcluded, "consumerGroupsExcluded", "c", []string{}, "excluded consumer group regexes from replication status")

	kafkaReplicationCmd.AddCommand(kafkaReplicationHealthStatusCmd)
	kafkaReplicationCmd.AddCommand(kafkaMonitorReplicationCmd)
}
