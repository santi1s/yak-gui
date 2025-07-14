package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	c "github.com/fatih/color"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type KafkaWorkerResetFlags struct {
	consumerGroupID          string
	consumerRuntime          string
	consumerRuntimeNamespace string
	topicsConsumedPrefix     string
	topicsProducedPrefix     string
	topicsSuffix             []string
	primaryKafkaCluster      string
	PrimaryAwsRegion         string
	secondaryKafkaCluster    string
	SecondaryAwsRegion       string
	databaseName             string
	databaseSchema           string
	workerType               string
	dbEngine                 string
	dryRun                   bool
}

var ProvidedKafkaWorkerResetFlags KafkaWorkerResetFlags

const consumerPlatform = "Kubernetes"

func shutdownConsumerRuntime(clusterName string, consumerRuntime string, consumerRuntimeNamespace string, dryRun bool) error {
	if consumerPlatform == "Kubernetes" {
		err := helper.ShutdownDeploymentInKubernetesWithDryRun(clusterName, consumerRuntime, consumerRuntimeNamespace, dryRun)
		if err != nil {
			log.Errorf("[%s] The consumer deployment %s running in ns %s has not been shutdown because of %v\n",
				clusterName, consumerRuntime, consumerRuntimeNamespace, err)
			confirmed := cli.AskConfirmation("\nCan you confirm that you introduced the right deployment name and namespace and that you want to continue despite the error?")
			if confirmed {
				log.Infof("[%s] We are assuming at this point that the consumer deployment is already deleted ...\n", clusterName)
				return nil
			}
		}
		return err
	}
	return nil
}

// Get the status of consumerGroupID on kafka cluster names clusterName
// do that until status of the consumerGroupID is equal to Empty
func waitForConsumerGroupEmptyStatus(clusterName string, awsRegion string, cfgFile string, consumerGroupID string, roleBasedAuth bool, dryRun bool) {
	for {
		state, err := helper.GetConsumerGroupState(clusterName, awsRegion, cfgFile, consumerGroupID, roleBasedAuth)
		if err != nil {
			log.Errorf("Couldn't retrieve the consumer group %s state, because of %v",
				consumerGroupID, err)
			panic(err)
		}
		cli.Printf("\nThe consumer group %s is in state : %s \n", consumerGroupID, state)
		if strings.Compare(state, "Empty") == 0 {
			break
		}
		if dryRun {
			cli.Printf("\n[DRY RUN] Wait until the consumer group %s state would be Empty\n", consumerGroupID)
			break
		}
		time.Sleep(5 * time.Second)
	}
}

// reset some topics offset on the consumer group called consumerGroupID on the kafka cluster named clusterName
func resetConsumerGroupOffsetOnSomeTopics(clusterName string, awsRegion string, cfgFile string, topics []string, consumerGroupID string, roleBasedAuth bool, dryRun bool) []error {
	if len(topics) == 0 {
		log.Errorln("No topic to reset consumer group offset on")
		return nil
	}
	cli.Printf("\n[%s] Visualize the consumer group %s on cluster %s\n", clusterName, consumerGroupID, clusterName)
	err := helper.VisualizeConsumerGroupOffsets(clusterName, awsRegion, cfgFile, consumerGroupID, topics, false, roleBasedAuth)
	if err != nil {
		panic(err)
	}

	cli.Printf("\n[%s] Reset topics offsets on the consumer group %s\n", clusterName, consumerGroupID)
	errors := helper.ResetConsumerGroupOffsets(clusterName, awsRegion, cfgFile, consumerGroupID, topics, roleBasedAuth, true, dryRun)

	cli.Printf("\n[%s] Visualize the consumer group %s on cluster %s\n", clusterName, consumerGroupID, clusterName)
	err = helper.VisualizeConsumerGroupOffsets(clusterName, awsRegion, cfgFile, consumerGroupID, topics, false, roleBasedAuth)
	if err != nil {
		panic(err)
	}
	return errors
}

// build a list of topics based on a topic prefix and a topics list called topicsSuffix
func constructTopics(topicPrefix string, topicsSuffix []string) []string {
	var constructedTopics []string
	if strings.Compare(topicPrefix, "") == 0 {
		return constructedTopics
	}
	for _, topic := range topicsSuffix {
		constructedTopics = append(constructedTopics, fmt.Sprintf("%s.%s", topicPrefix, topic))
	}
	return constructedTopics
}

// ask the operator to do some tasks and verifications before moving forward with the task
func operatorRequestForPretasks(primaryClusterName string, secondaryClusterName string, consumerGroupID string, consumerRuntime string, consumerRuntimeNamespace string, topicsConsumed []string, topicsProduced []string) {
	blue := c.New(c.FgBlue).SprintFunc()
	topicsProducedString := ""
	for _, topic := range topicsProduced {
		topicsProducedString += fmt.Sprintf("-%s\n", topic)
	}
	topicsConsumedString := ""
	for _, topic := range topicsConsumed {
		topicsConsumedString += fmt.Sprintf("-%s\n", topic)
	}
	cli.AskConfirmation(fmt.Sprintf("\n[%s][%s] Ensure that this is the consumer group id you would like to operate on %s", blue(primaryClusterName), blue(secondaryClusterName), consumerGroupID))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that the consumer runtime is the deployment %s running inside the namespace %s", blue(primaryClusterName), consumerRuntime, consumerRuntimeNamespace))
	cli.AskConfirmation(fmt.Sprintf("\n[%s][%s] Ensure that these topics are the ones that you want to reset the consumer group offset for : \n%v\n", blue(primaryClusterName), blue(secondaryClusterName), topicsConsumedString))
	cli.AskConfirmation(fmt.Sprintf("\n[%s][%s] Ensure that these topics are the ones that you want to delete from the primary and secondary clusters : \n%v\n", blue(primaryClusterName), blue(secondaryClusterName), topicsProducedString))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have suspended the the argocd app managing the consumer runtime ... \n> yak argocd suspend -a %s", blue(primaryClusterName), consumerRuntimeNamespace))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have suspended the cronjob that autorollout the consumer runtime ... \n> kubectl patch cronjobs.batch -n %s autodeploy-....  -p '{\"spec\" : {\"suspend\" : false }}'", blue(primaryClusterName), consumerRuntimeNamespace))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have suspended the argocd app managing mirrormaker2 ... \n> yak argocd suspend -a kafka", blue(secondaryClusterName)))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have deleted all mirrormaker2 deployments ... \n> kubectl delete deployments -n kafka mirrormaker2-mirror-source mirrormaker2-mirror-checkpoint mirrormaker2-mirror-heartbeat", blue(secondaryClusterName)))
}

// ask the operator to do some tasks and verifications before moving forward with the task
func operatorRequestForPosttasks(primaryClusterName string, secondaryClusterName string, consumerRuntimeNamespace string) {
	blue := c.New(c.FgBlue).SprintFunc()
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have unsuspended the the argocd app managing the consumer runtime ... \n> yak argocd unsuspend -a %s", blue(primaryClusterName), consumerRuntimeNamespace))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have unsuspended the cronjob that autorollout the consumer runtime ... \n> kubectl patch cronjobs.batch -n debezium autodeploy-....  -p '{\"spec\" : {\"suspend\" : true }}'", blue(primaryClusterName)))
	cli.AskConfirmation(fmt.Sprintf("\n[%s] Ensure that you have unsuspended the argocd app managing mirrormaker2 ... \n> yak argocd unsuspend -a kafka", blue(secondaryClusterName)))
}

func printMessageAndSleepSeconds(message string, seconds int) {
	log.Infoln(message)
	time.Sleep(time.Duration(seconds) * time.Second)
}

func deleteTopicsAndResetOffsets(clusterName string, awsRegion string, consumerGroupID string, topicsProduced []string, topicsConsumed []string, cfgFile string, roleBasedAuth bool, dryRun bool) ([]error, []error) {
	cli.LoglnAndSleep(3, fmt.Sprintln("Here starts operations on the kafka cluster: ", clusterName, "..."))

	cli.LoglnAndSleep(3, fmt.Sprintln("Deletion of topics on cluster ", clusterName, "..."))
	errorsDeletionTopics := helper.DeleteTopics(clusterName, awsRegion, cfgFile, topicsProduced, roleBasedAuth, true, dryRun)
	if errorsDeletionTopics != nil {
		log.Errorln(fmt.Sprintf("[%s]Errors while deleting topics : %v", clusterName, errorsDeletionTopics))
	}

	cli.LoglnAndSleep(3, fmt.Sprintln("Reset of offsets for consumer", consumerGroupID, "on cluster", clusterName, "..."))
	errorsResetOffsets := resetConsumerGroupOffsetOnSomeTopics(clusterName, awsRegion, cfgFile, topicsConsumed, consumerGroupID, roleBasedAuth, dryRun)
	if errorsResetOffsets != nil {
		log.Errorln(fmt.Sprintf("[%s]Errors while resetting topics on consumer group %s: %v", clusterName, consumerGroupID, errorsResetOffsets))
	}

	return errorsDeletionTopics, errorsResetOffsets
}

func extractOnlyExistantTopics(clusterName string, awsRegion string, cfgFile string, roleBasedAuth bool, topics []string) ([]string, error) {
	yellow := c.New(c.FgYellow).SprintFunc()
	green := c.New(c.FgGreen).SprintFunc()

	topicsExistant, err := helper.ExcludeNonExistentTopics(clusterName, awsRegion, cfgFile, topics, roleBasedAuth)
	if len(topics) == len(topicsExistant) {
		cli.Printf("\n[%s] All topics exist: %v\n", green(clusterName), topics)
	} else {
		cli.Printf("\n[%s] The topics that actually exist are: %v\n", yellow(clusterName), topicsExistant)
	}
	return topicsExistant, err
}

// Verify that the user is setting the right context
func executeVerifications(primaryClusterName string, primaryAwsRegion string, topicsConsumed []string, topicsProduced []string, cfgFile string, roleBasedAuth bool) ([]string, []string) {
	cli.Printf("\n[%s] Verify that you're setting the right kubernetes context\n", primaryClusterName)
	context, err := helper.GetKubernetesCurrentContext()
	if err != nil {
		panic(err)
	}

	log.Infoln("The current context is:", context)
	if context == primaryClusterName {
		log.Infoln("You're setting the right kubernetes context")
	} else {
		if context == fmt.Sprintf("teleport.doctolib.net-%s", primaryClusterName) {
			log.Infoln("You're setting the right kubernetes context and you're using teleport")
		} else {
			log.Fatalln("Your current context is not correct")
			panic(errors.New("Please set your kubernetes context correctly"))
		}
	}
	// Verify that all of these topics exist on the primary cluster
	cli.Printf("\n[%s] Verify that all consumed topics communicated exist on the cluster: %v", primaryClusterName, topicsConsumed)
	topicsConsumedOnlyExistent, err := extractOnlyExistantTopics(primaryClusterName, primaryAwsRegion, cfgFile, roleBasedAuth, topicsConsumed)
	if err != nil {
		panic(err)
	}

	cli.Printf("\n[%s] Verify that all produced topics communicated exist on the cluster: %v", primaryClusterName, topicsProduced)
	topicsProducedOnlyExistent, err := extractOnlyExistantTopics(primaryClusterName, primaryAwsRegion, cfgFile, roleBasedAuth, topicsProduced)
	if err != nil {
		panic(err)
	}
	return topicsConsumedOnlyExistent, topicsProducedOnlyExistent
}

// workerReset will delete topicsProduced by the consumer consumerGroupID
// and reset the offset of consumerGroupID to 0 on the topicsConsumed
// topicsProduced and topicsConsumed are two topics list constructed based on
// topicsProducedPrefix, topicsConsumedPrefix and topicsSuffix
func workerReset(cmd *cobra.Command, _ []string) error {
	var errTopicsSecondary []error
	var errResetOffsetsSecondary []error

	primaryClusterName := ProvidedKafkaWorkerResetFlags.primaryKafkaCluster
	primaryAwsRegion := ProvidedKafkaWorkerResetFlags.PrimaryAwsRegion
	secondaryClusterName := ProvidedKafkaWorkerResetFlags.secondaryKafkaCluster
	secondaryAwsRegion := ProvidedKafkaWorkerResetFlags.SecondaryAwsRegion
	roleBasedAuth := ProvidedKafkaFlags.roleBasedAuth
	cfgFile := ProvidedKafkaFlags.cfgFile
	topicsConsumedPrefix := ProvidedKafkaWorkerResetFlags.topicsConsumedPrefix
	topicsProducedPrefix := ProvidedKafkaWorkerResetFlags.topicsProducedPrefix
	topicsSuffix := ProvidedKafkaWorkerResetFlags.topicsSuffix
	consumerGroupID := ProvidedKafkaWorkerResetFlags.consumerGroupID
	consumerRuntime := ProvidedKafkaWorkerResetFlags.consumerRuntime
	consumerRuntimeNamespace := ProvidedKafkaWorkerResetFlags.consumerRuntimeNamespace
	dbName := ProvidedKafkaWorkerResetFlags.databaseName
	dbEngine := ProvidedKafkaWorkerResetFlags.dbEngine
	dbSchema := ProvidedKafkaWorkerResetFlags.databaseSchema
	workerType := ProvidedKafkaWorkerResetFlags.workerType
	dryRun := ProvidedKafkaWorkerResetFlags.dryRun

	// Load primary cluster region
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}
	if !cmd.Flags().Changed("primary-aws-region") {
		primaryAwsRegion = cfg.Region
	}

	// based on our naming conventions, we can construct
	// consumerRuntime, topicsConsumedPrefix and topicsProducedPrefix
	// relying on the worker's type (anonymizer|deleted)
	// relying on the db schema
	// relying on the db name
	// relyinh on the db engine as well
	if consumerRuntime == "" {
		consumerRuntime = fmt.Sprintf("kafka-worker-%s-%s-%s", workerType, dbName, dbEngine)
	}

	if topicsConsumedPrefix == "" {
		topicsConsumedPrefix = fmt.Sprintf("%s.data.%s.%s.%s", consumerRuntimeNamespace, dbName, dbEngine, dbSchema)
	}

	if topicsProducedPrefix == "" {
		topicsProducedPrefix = fmt.Sprintf("%s.%s.%s.%s.%s", consumerRuntimeNamespace, workerType, dbName, dbEngine, dbSchema)
	}

	if consumerGroupID == "" {
		consumerGroupID = fmt.Sprintf("%s.kafka-worker-%s-%s-%s", consumerRuntimeNamespace, workerType, dbName, dbEngine)
	}

	topicsConsumed := constructTopics(topicsConsumedPrefix, topicsSuffix)
	topicsProduced := constructTopics(topicsProducedPrefix, topicsSuffix)

	if dryRun {
		cli.LoglnAndSleep(3, "DRY RUN MODE ACTIVATED .... ")
	} else {
		cli.LoglnAndSleep(3, "DRY RUN DISABLED .... ")
	}

	// Verify that all concerned topics exist on the primary cluster and cleanup the non-existent ones
	topicsConsumedOnlyExistent, topicsProducedOnlyExistent := executeVerifications(primaryClusterName, primaryAwsRegion, topicsConsumed, topicsProduced, cfgFile, roleBasedAuth)

	// Ask the operator to run/verify some tasks before the beginning of the script
	operatorRequestForPretasks(primaryClusterName, secondaryClusterName, consumerGroupID, consumerRuntime, consumerRuntimeNamespace, topicsConsumedOnlyExistent, topicsProducedOnlyExistent)

	// shutdown the consumer runtime on the active cluster
	err = shutdownConsumerRuntime(primaryClusterName, consumerRuntime, consumerRuntimeNamespace, dryRun)
	if err != nil {
		panic(err)
	}

	// wait for all consumer group members to disconnect from kafka
	waitForConsumerGroupEmptyStatus(primaryClusterName, primaryAwsRegion, cfgFile, consumerGroupID, roleBasedAuth, dryRun)

	errTopicsPrimary, errResetOffsetsPrimary := deleteTopicsAndResetOffsets(primaryClusterName, primaryAwsRegion, consumerGroupID, topicsProducedOnlyExistent, topicsConsumedOnlyExistent, cfgFile, roleBasedAuth, dryRun)
	if strings.Compare(secondaryClusterName, "") != 0 {
		errTopicsSecondary, errResetOffsetsSecondary = deleteTopicsAndResetOffsets(secondaryClusterName, secondaryAwsRegion, consumerGroupID, topicsProducedOnlyExistent, topicsConsumedOnlyExistent, cfgFile, roleBasedAuth, dryRun)
	}
	if (errTopicsPrimary != nil || len(errTopicsPrimary) > 0) ||
		(errResetOffsetsPrimary != nil || len(errResetOffsetsPrimary) > 0) ||
		(errTopicsSecondary != nil || len(errTopicsSecondary) > 0) ||
		(errResetOffsetsSecondary != nil || len(errResetOffsetsSecondary) > 0) {
		log.Errorln("Done with errors, analyse the previous logs to understand what happend")
		return errors.New("Errors")
	}
	operatorRequestForPosttasks(primaryClusterName, secondaryClusterName, consumerRuntimeNamespace)
	log.Infoln("DONE Successfully!")
	return nil
}

var kafkaWorkerResetCmd = &cobra.Command{
	Use:   "worker-reset",
	Short: "Reset some topics offset for a kafka worker (consumer group) & delete the corresponding produced topics on both primary and secondary kafka clusters",
	RunE:  workerReset,
	Args:  cobra.ExactArgs(0),
	Example: `(long format) > yak kafka worker-reset -g debezium.kafka-worker-anonymized-mdp-postgres -C kafka-worker-anonymized-mdp-postgres -p staging-aws-fr-par-1 -P eu-west-3 -s staging-aws-de-fra-1 -S eu-central-1 -R debezium.data.mdp.postgres.public -D debezium.anonymized.mdp.postgres.public -x conditions -x exams
(short format) > yak kafka worker-reset -p staging-aws-fr-par-1 -s staging-aws-de-fra-1 -S eu-central-1 -d mdp -x conditions -x exams`,
}

func init() {
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.consumerGroupID, "consumer-group-id", "g", "", "the consumer group id on kafka")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.consumerRuntime, "consumer-runtime", "C", "", "consumer runtime name")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.consumerRuntimeNamespace, "consumer-runtime-namespace", "n", "debezium", "consumer runtime namespace")

	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.databaseName, "db-name", "d", "", "database name")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.databaseSchema, "db-schema", "m", "public", "database schema name")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.dbEngine, "db-engine", "e", "postgres", "database engine")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.workerType, "worker-type", "t", "anonymized", "worker type, example: anonymized")

	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.topicsConsumedPrefix, "topics-consumed-prefix", "R", "", "the shared prefix of the topics consumed by the consumer group and that need to be reset")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.topicsProducedPrefix, "topics-produced-prefix", "D", "", "the shared prefix of the topics produced by the consumer group and that need to be deleted")

	kafkaWorkerResetCmd.PersistentFlags().StringArrayVarP(&ProvidedKafkaWorkerResetFlags.topicsSuffix, "topics-suffix", "x", []string{}, "the topics suffix (table names) of topics")

	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.primaryKafkaCluster, "primary-cluster", "p", "", "kafka primary cluster name, example: staging-aws-de-fra-1")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.secondaryKafkaCluster, "secondary-cluster", "s", "", "kafka secondary cluster name")

	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.PrimaryAwsRegion, "primary-aws-region", "P", "", "primary aws region name, example: eu-west-3")
	kafkaWorkerResetCmd.PersistentFlags().StringVarP(&ProvidedKafkaWorkerResetFlags.SecondaryAwsRegion, "secondary-aws-region", "S", "", "secondary aws region name, example: eu-central-1")
	kafkaWorkerResetCmd.PersistentFlags().BoolVarP(&ProvidedKafkaWorkerResetFlags.dryRun, "dry-run", "r", false, "execute a dry run of this task")

	_ = kafkaWorkerResetCmd.MarkFlagRequired("primary-cluster")
	_ = kafkaWorkerResetCmd.MarkFlagRequired("primary-aws-region")
	_ = kafkaWorkerResetCmd.MarkFlagRequired("topics-suffix")
	_ = kafkaWorkerResetCmd.MarkFlagRequired("consumer-group-id")
	_ = kafkaWorkerResetCmd.MarkFlagRequired("db-name")
	kafkaWorkerResetCmd.MarkFlagsMutuallyExclusive("db-name", "consumer-runtime")
	kafkaWorkerResetCmd.MarkFlagsMutuallyExclusive("db-schema", "consumer-runtime")
	kafkaWorkerResetCmd.MarkFlagsMutuallyExclusive("worker-type", "consumer-runtime")
	kafkaWorkerResetCmd.MarkFlagsMutuallyExclusive("db-engine", "consumer-runtime")
}
