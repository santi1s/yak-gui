package kafka

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/santi1s/yak/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

type KafkaMonitorReplicationFlags struct {
	prometheusPort int
	prometheusPath string
}

var (
	// https://prometheus.io/docs/practices/naming/#metric-names
	unreplicatedTopics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_unreplicated_topics_total",
			Help: "Count of topics that exist on the primary kafka cluster and not on the secondary cluster",
		},
		[]string{"primaryCluster", "secondaryCluster"},
	)

	unreplicatedConsumerGroups = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_unreplicated_consumer_groups_total",
			Help: "Count of consumer groups present of the primary kafka cluster and not on the secondary cluster",
		},
		[]string{"primaryCluster", "secondaryCluster"},
	)

	orphanTopics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_orphan_topics_total",
			Help: "Count of topics that exist on the secondary kafka cluster and not on the primary cluster",
		},
		[]string{"primaryCluster", "secondaryCluster"},
	)

	orphanConsumerGroups = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_orphan_consumer_groups_total",
			Help: "Count of consumer groups present of the secondary kafka cluster and not on the primary cluster",
		},
		[]string{"primaryCluster", "secondaryCluster"},
	)

	mutexForUpdatingPromMetrics sync.Mutex

	ProvidedKafkaMonitorReplicationFlags KafkaMonitorReplicationFlags

	kafkaMonitorReplicationCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Expose prometheus metrics around kafka replication status",
		RunE:  prometheusReplicationStatus,
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
)

// updateKafkaReplicationMetrics updates continuously prometheus metrics related to kafka topics and consumerGroup offset replication
func updateKafkaReplicationMetrics(primaryClusterName string, secondaryClusterName string, primaryAwsRegion string, secondaryAwsRegion string, excludedTopics []string, excludedConsumerGroups []string, cfgFile string, roleBasedAuth bool) {
	for {
		topicsInPrimaryButNotInSecondary, topicsInSecondaryButNotInPrimary, err := CompareTopics(primaryClusterName, secondaryClusterName, primaryAwsRegion, secondaryAwsRegion, excludedTopics, cfgFile, roleBasedAuth)
		if err != nil {
			log.Fatal("Unable to grab topics from one of the kafka clusters :", err)
		}
		consumerGroupsInPrimaryButNotInSecondary, consumerGroupsInSecondaryButNotInPrimary, err := CompareConsumerGroups(primaryClusterName, secondaryClusterName, primaryAwsRegion, secondaryAwsRegion, excludedConsumerGroups, cfgFile, roleBasedAuth)
		if err != nil {
			log.Fatal("Unable to grab consumer groups from one of the kafka clusters :", err)
		}

		mutexForUpdatingPromMetrics.Lock()
		unreplicatedTopics.WithLabelValues(primaryClusterName, secondaryClusterName).Set(float64(len(topicsInPrimaryButNotInSecondary)))
		mutexForUpdatingPromMetrics.Unlock()

		mutexForUpdatingPromMetrics.Lock()
		orphanTopics.WithLabelValues(primaryClusterName, secondaryClusterName).Set(float64(len(topicsInSecondaryButNotInPrimary)))
		mutexForUpdatingPromMetrics.Unlock()

		mutexForUpdatingPromMetrics.Lock()
		unreplicatedConsumerGroups.WithLabelValues(primaryClusterName, secondaryClusterName).Set(float64(len(consumerGroupsInPrimaryButNotInSecondary)))
		mutexForUpdatingPromMetrics.Unlock()

		mutexForUpdatingPromMetrics.Lock()
		orphanConsumerGroups.WithLabelValues(primaryClusterName, secondaryClusterName).Set(float64(len(consumerGroupsInSecondaryButNotInPrimary)))
		mutexForUpdatingPromMetrics.Unlock()

		// Sleep for a period of 1 min before updating metrics again
		sleepDuration := 1 * time.Minute
		time.Sleep(sleepDuration)
	}
}

func prometheusReplicationStatus(_ *cobra.Command, _ []string) error {
	prometheus.MustRegister(unreplicatedTopics)
	prometheus.MustRegister(unreplicatedConsumerGroups)
	prometheus.MustRegister(orphanTopics)
	prometheus.MustRegister(orphanConsumerGroups)
	// HTTP handler to expose Prometheus metrics
	http.Handle(ProvidedKafkaMonitorReplicationFlags.prometheusPath, promhttp.Handler())
	primaryCluster := ProvidedKafkaReplicationFlags.PrimaryCluster
	primaryAwsRegion := ProvidedKafkaReplicationFlags.PrimaryAwsRegion
	secondaryCluster := ProvidedKafkaReplicationFlags.SecondaryCluster
	secondaryAwsRegion := ProvidedKafkaReplicationFlags.SecondaryAwsRegion
	excludedTopics := ProvidedKafkaReplicationFlags.topicsRegexExcluded
	excludedConsumerGroups := ProvidedKafkaReplicationFlags.consumerGroupsRegexExcluded
	roleBasedAuth := ProvidedKafkaFlags.roleBasedAuth
	cfgFile := ProvidedKafkaFlags.cfgFile
	go updateKafkaReplicationMetrics(primaryCluster, secondaryCluster, primaryAwsRegion, secondaryAwsRegion, excludedTopics, excludedConsumerGroups, cfgFile, roleBasedAuth)
	// Start the HTTP server
	port := ProvidedKafkaMonitorReplicationFlags.prometheusPort
	cli.Printf("Server listening on :%d\n", port)
	err := http.ListenAndServe(":"+fmt.Sprint(port), nil) //#nosec G114
	if err != nil {
		cli.Println("Error starting server:", err)
		return err
	}
	return nil
}

func init() {
	kafkaMonitorReplicationCmd.PersistentFlags().IntVarP(&ProvidedKafkaMonitorReplicationFlags.prometheusPort, "port", "r", 4000, "port for prometheus server")
	kafkaMonitorReplicationCmd.PersistentFlags().StringVarP(&ProvidedKafkaMonitorReplicationFlags.prometheusPath, "path", "m", "/metrics", "metrics path for prometheus server")
}
