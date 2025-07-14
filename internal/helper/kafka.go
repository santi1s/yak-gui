package helper

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	mskSigner "github.com/aws/aws-msk-iam-sasl-signer-go/signer"
	"github.com/santi1s/yak/cli"
	log "github.com/sirupsen/logrus"

	"github.com/IBM/sarama"
	"github.com/birdayz/kaf/pkg/config"
)

const (
	IAMAuthVersion          = "2020_10_22"
	tabwriterMinWidth       = 6
	tabwriterMinWidthNested = 2
	tabwriterWidth          = 4
	tabwriterPadding        = 3
	tabwriterPadChar        = ' '
	tabwriterFlags          = 0
)

var (
	outWriter io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr
)

type partitionOffset struct {
	highWatermark int64
	offset        int64
}

type PartitionAssignment struct {
	partition int32
	offset    int64
}

type AWSMSKIAMConfig struct {
	Region string
	Expiry time.Duration
}

type MSKAccessTokenProvider struct {
	AwsRegion     string
	RoleBasedAuth bool
}

type resetHandler struct {
	topic            string
	partitionOffsets map[int32]int64
	offset           int64
	client           sarama.Client
	group            string
	mskProvider      MSKAccessTokenProvider
	clusterConfig    *config.Cluster
}

func (r *resetHandler) Setup(s sarama.ConsumerGroupSession) error {
	req := &sarama.OffsetCommitRequest{
		Version:                 1,
		ConsumerGroup:           r.group,
		ConsumerGroupGeneration: s.GenerationID(),
		ConsumerID:              s.MemberID(),
	}

	for p, o := range r.partitionOffsets {
		req.AddBlock(r.topic, p, o, 0, "")
	}
	br, err := r.client.Coordinator(r.group)
	if err != nil {
		return err
	}
	_ = br.Open(getConfig(r.clusterConfig, r.mskProvider.AwsRegion, r.mskProvider.RoleBasedAuth))
	_, err = br.CommitOffset(req)
	if err != nil {
		return err
	}
	return nil
}

func (r *resetHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	return nil
}

func (r *resetHandler) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	return nil
}

func (t *MSKAccessTokenProvider) Token() (*sarama.AccessToken, error) {
	var token string
	var err error
	region := &t.AwsRegion
	if t.RoleBasedAuth {
		roleArn := os.Getenv("AWS_ROLE_ARN")
		sessionName := strings.Split(roleArn, "/")[1]
		token, _, err = mskSigner.GenerateAuthTokenFromRole(context.TODO(), *region, roleArn, sessionName)
	} else {
		token, _, err = mskSigner.GenerateAuthToken(context.TODO(), *region)
	}
	if err != nil {
		log.Fatal("Cannot generate AWS MSK Access token", err)
		return nil, err
	}
	log.Debugln("Successfully AWS MSK Access token generated")
	return &sarama.AccessToken{Token: token}, err
}

func getClusterAdmin(clusterName string, awsRegion string, cfgFile string, roleBasedAuth bool) sarama.ClusterAdmin {
	var err error
	var currentCluster *config.Cluster
	cfg, err := config.ReadConfig(cfgFile)
	if err != nil {
		log.Fatal("Invalid config: ", err)
	}
	cfg.ClusterOverride = clusterName
	cluster := cfg.ActiveCluster()
	if cluster != nil {
		// Use active cluster from config
		currentCluster = cluster
	} else {
		log.Fatal("Cluster not found : ", clusterName)
	}
	clusterAdmin, err := sarama.NewClusterAdmin(currentCluster.Brokers, getConfig(cluster, awsRegion, roleBasedAuth))
	if err != nil {
		fmt.Println(currentCluster.Brokers)
		fmt.Println(cluster)
		fmt.Println(awsRegion)
		log.Fatal("Unable to get cluster admin: ", err)
	}
	log.Debugln("ClusterAdmin created for cluster:", clusterName)
	return clusterAdmin
}

func getClient(clusterName string, awsRegion string, cfgFile string, roleBasedAuth bool) sarama.Client {
	var err error
	var currentCluster *config.Cluster
	cfg, err := config.ReadConfig(cfgFile)
	if err != nil {
		log.Fatal("Invalid config: ", err)
	}
	cfg.ClusterOverride = clusterName
	cluster := cfg.ActiveCluster()
	if cluster != nil {
		// Use active cluster from config
		currentCluster = cluster
	} else {
		log.Fatal("Cluster not found : ", clusterName)
	}
	client, err := sarama.NewClient(currentCluster.Brokers, getConfig(cluster, awsRegion, roleBasedAuth))
	if err != nil {
		log.Errorf("Unable to get client: %v\n", err)
	}
	return client
}

func getConfig(cluster *config.Cluster, awsRegion string, roleBasedAuth bool) *sarama.Config {
	config := sarama.NewConfig()
	var tlsInsecure bool
	if cluster.Version != "" {
		parsedVersion, err := sarama.ParseKafkaVersion(cluster.Version)
		if err != nil {
			log.Fatal("Unable to parse Kafka version: ", err)
		}
		config.Version = parsedVersion
	}
	config.Net.SASL.Enable = true
	if cluster.SASL.Mechanism == "AWS_MSK_IAM" {
		config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		config.Net.SASL.TokenProvider = &MSKAccessTokenProvider{AwsRegion: awsRegion, RoleBasedAuth: roleBasedAuth}
	}

	if cluster.SASL.Mechanism == "PLAIN" {
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		config.Net.SASL.User = cluster.SASL.Username
		config.Net.SASL.Password = cluster.SASL.Password
	}
	if cluster.TLS != nil {
		tlsInsecure = cluster.TLS.Insecure
	} else {
		tlsInsecure = false
	}
	tlsConfig := tls.Config{InsecureSkipVerify: tlsInsecure, MinVersion: 0x0303} // #nosec G402
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tlsConfig
	log.Debugln("Here's the sarama config", config)
	return config
}

// GetTopics reports the topics inside the kafka cluster named clusterName.
func GetTopics(clusterName string, awsRegion string, cfgFile string, roleBasedAuth bool) (map[string]sarama.TopicDetail, error) {
	clusterAdmin := getClusterAdmin(clusterName, awsRegion, cfgFile, roleBasedAuth)
	topics, err := clusterAdmin.ListTopics()
	if err != nil {
		log.Fatal("Unable to list topics: ", err)
	}
	log.Debugln("Successfully grabbed topics for cluster:", clusterName)
	clusterAdmin.Close()
	return topics, nil
}

func deleteTopic(clusterName string, awsRegion string, cfgFile string, topic string, roleBasedAuth bool) error {
	clusterAdmin := getClusterAdmin(clusterName, awsRegion, cfgFile, roleBasedAuth)
	err := clusterAdmin.DeleteTopic(topic)
	clusterAdmin.Close()
	return err
}

// GetConsumerGroups reports the consumer groups inside the kafka cluster named clusterName.
func GetConsumerGroups(clusterName string, awsRegion string, cfgFile string, typesIncluded []string, roleBasedAuth bool) ([]string, error) {
	clusterAdmin := getClusterAdmin(clusterName, awsRegion, cfgFile, roleBasedAuth)
	consumerGroupNames, err := clusterAdmin.ListConsumerGroups()
	if err != nil {
		log.Fatal("Unable to list consumer groups: ", err)
	}
	groupList := make([]string, 0, len(consumerGroupNames))
	for grp := range consumerGroupNames {
		groupList = append(groupList, grp)
	}
	if len(typesIncluded) == 0 {
		sort.Strings(groupList)
		return groupList, nil
	}
	groupsDescription, err := clusterAdmin.DescribeConsumerGroups(groupList)
	if err != nil {
		return []string{}, err
	}
	log.Debugln("Successfully grabbed consumer groups descriptions")
	filteredGroupList := make([]string, 0, len(consumerGroupNames))
	for _, groupDescription := range groupsDescription {
		for _, typeConsumerGroup := range typesIncluded {
			if groupDescription.ProtocolType == typeConsumerGroup || groupDescription.ProtocolType == "" {
				filteredGroupList = append(filteredGroupList, groupDescription.GroupId)
				break
			}
		}
	}
	log.Debugln("Consumer group list has been successfully filtered")
	sort.Strings(filteredGroupList)
	clusterAdmin.Close()
	return filteredGroupList, nil
}

func getClusterConfig(clusterName string, cfgFile string) (*config.Cluster, error) {
	cfg, err := config.ReadConfig(cfgFile)
	if err != nil {
		log.Errorln("Invalid config: ", err)
		return nil, err
	}
	cfg.ClusterOverride = clusterName
	cluster := cfg.ActiveCluster()
	if cluster == nil {
		log.Errorf("[%s] Cluster not found \n", clusterName)
		return nil, err
	}
	return cluster, nil
}

func ResetConsumerGroupOffsets(clusterName string, awsRegion string, cfgFile string, consumerGroupID string, topics []string, roleBasedAuth bool, askConfirmation bool, dryRun bool) []error {
	assignedPartitionOffsets := make(map[int32]int64)
	var errors []error
	var confirmed = true
	if askConfirmation {
		topicsString := ""
		for _, topic := range topics {
			topicsString += fmt.Sprintf("-%s\n", topic)
		}
		confirmed = cli.AskConfirmation(fmt.Sprintf("\nDo you want to set offset to 0 on group %s for all topic partitions \n%s", consumerGroupID, topicsString))
	}
	if !confirmed {
		log.Warn(fmt.Sprintf("[%s][Rejected] No reset will be done on %s for any topic partitions \n%s", clusterName, consumerGroupID, topics))
		return nil
	}
	if len(topics) == 0 {
		log.Warnln("No topics to reset consumer group offsets on")
		return nil
	}
	client := getClient(clusterName, awsRegion, cfgFile, roleBasedAuth)
	group, offsets, err := GetConsumerGroupOffsetsOnTopics(clusterName, awsRegion, cfgFile, consumerGroupID, topics, false, roleBasedAuth)
	if err != nil {
		errors = append(errors, err)
		return errors
	}
	clusterConfig, err := getClusterConfig(clusterName, cfgFile)
	if err != nil {
		errors = append(errors, err)
		return errors
	}
	g, err := sarama.NewConsumerGroupFromClient(group.GroupId, client)
	if err != nil {
		log.Errorf("Failed to create consumer group: %v\n", err)
		errors = append(errors, err)
		return errors
	}
	for topic, partitionOffsets := range offsets {
		assignments := make([]PartitionAssignment, len(partitionOffsets))
		for partition := range partitionOffsets {
			assignments = append(assignments, PartitionAssignment{partition: partition, offset: 0})
		}
		for _, assign := range assignments {
			assignedPartitionOffsets[assign.partition] = assign.offset
		}
		if dryRun {
			log.Infof("[DRY RUN] the offset of topic %s would be reset to 0 on all its partitions on consumer group %s\n", topic, group.GroupId)
		} else {
			err = g.Consume(context.Background(), []string{topic}, &resetHandler{
				topic:            topic,
				partitionOffsets: assignedPartitionOffsets,
				client:           client,
				group:            group.GroupId,
				mskProvider:      MSKAccessTokenProvider{AwsRegion: awsRegion, RoleBasedAuth: roleBasedAuth},
				clusterConfig:    clusterConfig,
			})
			if err != nil {
				log.Errorf("Failed to commit offset: %v\n", err)
				errors = append(errors, err)
			}
			log.Infof("Successfully committed offsets on topic %s for consumer group %s to %v.\n", topic, group.GroupId, assignedPartitionOffsets)
		}
	}
	closeErr := g.Close()
	if closeErr != nil {
		log.Warnf("Failed to close consumer group: %v\n", closeErr)
	}
	return errors
}

func getHighWatermarksTopic(clusterName string, awsRegion string, topic string, partitions []int32, cfgFile string, roleBasedAuth bool) map[int32]int64 {
	var watermarks map[int32]int64
	client := getClient(clusterName, awsRegion, cfgFile, roleBasedAuth)
	leaders := make(map[*sarama.Broker][]int32)

	for _, partition := range partitions {
		leader, err := client.Leader(topic, partition)
		if err != nil {
			log.Errorf("Unable to get available offsets for partition without leader. Topic %s Partition %d, Error: %s \n", topic, partition, err)
		}
		leaders[leader] = append(leaders[leader], partition)
	}
	wg := sync.WaitGroup{}
	wg.Add(len(leaders))

	results := make(chan map[int32]int64, len(leaders))

	for leader, partitions := range leaders {
		req := &sarama.OffsetRequest{
			Version: int16(1),
		}

		for _, partition := range partitions {
			req.AddBlock(topic, partition, int64(-1), int32(0))
		}

		// Query distinct brokers in parallel
		go func(leader *sarama.Broker, req *sarama.OffsetRequest) {
			resp, err := leader.GetAvailableOffsets(req)
			if err != nil {
				log.Errorf("Unable to get available offsets: %v %v %v\n", req, resp, err)
				wg.Done()
			}
			watermarksFromLeader := make(map[int32]int64)
			for partition, block := range resp.Blocks[topic] {
				watermarksFromLeader[partition] = block.Offset
			}
			leader.Close()
			results <- watermarksFromLeader
			wg.Done()
		}(leader, req)
	}

	wg.Wait()
	close(results)

	watermarks = make(map[int32]int64)
	for resultMap := range results {
		for partition, offset := range resultMap {
			watermarks[partition] = offset
		}
	}
	client.Close()
	return watermarks
}

func GetConsumerGroupOffsetsOnTopics(clusterName string, awsRegion string, cfgFile string, consumerGroupID string, topics []string, allTopics bool, roleBasedAuth bool) (*sarama.GroupDescription, map[string](map[int32]partitionOffset), error) {
	clusterAdmin := getClusterAdmin(clusterName, awsRegion, cfgFile, roleBasedAuth)
	groups, err := clusterAdmin.DescribeConsumerGroups([]string{consumerGroupID})
	if err != nil {
		log.Errorf("Unable to describe consumer groups: %v\n", err)
		return nil, nil, err
	}
	if len(groups) == 0 {
		log.Errorf("Did not receive expected describe consumergroup result\n")
	}
	group := groups[0]

	if group.State == "Dead" {
		fmt.Printf("Group %s not found.\n", consumerGroupID)
		return group, nil, nil
	}
	offsetAndMetadata, err := clusterAdmin.ListConsumerGroupOffsets(consumerGroupID, nil)
	clusterAdmin.Close()
	if err != nil {
		log.Errorf("Failed to fetch group offsets: %v\n", err)
		return group, nil, err
	}

	allTopicsOffset := make([]string, 0, len(offsetAndMetadata.Blocks))
	for k := range offsetAndMetadata.Blocks {
		allTopicsOffset = append(allTopicsOffset, k)
	}
	sort.Strings(allTopicsOffset)

	filteredTopicsOffset := make(map[string](map[int32]partitionOffset))

	for _, topic := range allTopicsOffset {
		var found bool
		partitions := offsetAndMetadata.Blocks[topic]
		if allTopics {
			filteredTopicsOffset[topic] = make(map[int32]partitionOffset)
			found = true
		} else {
			for _, topicToShow := range topics {
				if topic == topicToShow {
					filteredTopicsOffset[topic] = make(map[int32]partitionOffset)
					found = true
				}
			}
		}
		if !found {
			continue
		}
		var p []int32
		for partition := range partitions {
			p = append(p, partition)
		}
		sort.Slice(p, func(i, j int) bool {
			return p[i] < p[j]
		})
		wms := getHighWatermarksTopic(clusterName, awsRegion, topic, p, cfgFile, roleBasedAuth)
		for _, partition := range p {
			filteredTopicsOffset[topic][partition] = partitionOffset{highWatermark: wms[partition], offset: partitions[partition].Offset}
		}
	}
	return group, filteredTopicsOffset, nil
}

func GetConsumerGroupLag(clusterName string, awsRegion string, cfgFile string, consumer string, topics []string, allTopics bool, roleBasedAuth bool) (map[string]map[int32]int64, error) {
	_, offsets, err := GetConsumerGroupOffsetsOnTopics(clusterName, awsRegion, cfgFile, consumer, topics, allTopics, roleBasedAuth)
	if err != nil {
		return nil, err
	}

	lags := make(map[string]map[int32]int64)

	for topic, partitions := range offsets {
		lags[topic] = make(map[int32]int64)
		for partition, partitionOffset := range partitions {
			lag := (partitionOffset.highWatermark - partitionOffset.offset)
			lags[topic][partition] = lag
		}
	}
	return lags, nil
}

func VisualizeConsumerGroupOffsets(clusterName string, awsRegion string, cfgFile string, consumerGroupID string, topics []string, allTopics bool, roleBasedAuth bool) error {
	group, offsets, err := GetConsumerGroupOffsetsOnTopics(clusterName, awsRegion, cfgFile, consumerGroupID, topics, allTopics, roleBasedAuth)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(outWriter, tabwriterMinWidth, tabwriterWidth, tabwriterPadding, tabwriterPadChar, tabwriterFlags)
	fmt.Fprintf(w, "Group ID:\t%v\n", group.GroupId)
	fmt.Fprintf(w, "State:\t%v\n", group.State)
	fmt.Fprintf(w, "Protocol:\t%v\n", group.Protocol)
	fmt.Fprintf(w, "Protocol Type:\t%v\n", group.ProtocolType)
	fmt.Fprintf(w, "Offsets:\t\n")

	w.Flush()
	w.Init(outWriter, tabwriterMinWidthNested, 4, 2, tabwriterPadChar, tabwriterFlags)

	for topic, partitions := range offsets {
		fmt.Fprintf(w, "\t%v:\n", topic)
		fmt.Fprintf(w, "\t\tPartition\tGroup Offset\tHigh Watermark\tLag\t\n")
		fmt.Fprintf(w, "\t\t---------\t------------\t--------------\t---\n")
		lagSum := 0
		offsetSum := 0
		for partition, partitionOffset := range partitions {
			lag := (partitionOffset.highWatermark - partitionOffset.offset)
			lagSum += int(lag)
			offset := partitionOffset.offset
			offsetSum += int(offset)
			fmt.Fprintf(w, "\t\t%v\t%v\t%v\t%v\n", partition, partitionOffset.offset, partitionOffset.highWatermark, lag)
		}
		fmt.Fprintf(w, "\t\tTotal\t%d\t\t%d\t\n", offsetSum, lagSum)
	}
	w.Flush()
	return nil
}

func GetConsumerGroupState(clusterName string, awsRegion string, cfgFile string, consumerGroupID string, roleBasedAuth bool) (string, error) {
	clusterAdmin := getClusterAdmin(clusterName, awsRegion, cfgFile, roleBasedAuth)
	groupsDescription, err := clusterAdmin.DescribeConsumerGroups([]string{consumerGroupID})
	if err != nil {
		return "", err
	}
	return groupsDescription[0].State, nil
}

func DeleteTopics(clusterName string, awsRegion string, cfgFile string, topics []string, roleBasedAuth bool, askConfirmation bool, dryRun bool) []error {
	var errors []error
	if len(topics) == 0 {
		log.Errorln("No topic to delete")
		return nil
	}
	confirmed := true
	if askConfirmation {
		topicsString := ""
		for _, topic := range topics {
			topicsString += fmt.Sprintf("-%s\n", topic)
		}
		confirmed = cli.AskConfirmation(fmt.Sprintf("\nDo you want to delete on cluster %s the following topics \n%s", clusterName, topicsString))
	}
	if confirmed || !askConfirmation {
		for _, topic := range topics {
			if dryRun {
				cli.Printf("\n[DRY RUN] The topic %s would be deleted on cluster %s\n", topic, clusterName)
			} else {
				err := deleteTopic(clusterName, awsRegion, cfgFile, topic, roleBasedAuth)
				if err != nil {
					errors = append(errors, err)
				} else {
					cli.Printf("\n[%s]The topic %s is deleted from cluster\n", clusterName, topic)
				}
			}
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

func ExcludeNonExistentTopics(clusterName string, awsRegion string, cfgFile string, topics []string, roleBasedAuth bool) ([]string, error) {
	var topicsExistant []string
	topicDetails, err := GetTopics(clusterName, awsRegion, cfgFile, roleBasedAuth)
	if err != nil {
		return topicsExistant, err
	}
	for _, topic := range topics {
		for topicNameInCluster := range topicDetails {
			if strings.Compare(topic, topicNameInCluster) == 0 {
				topicsExistant = append(topicsExistant, topic)
				break
			}
		}
	}
	return topicsExistant, nil
}
