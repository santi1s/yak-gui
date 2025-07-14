package kafka

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

type KafkaMonitorConsumerLagFlags struct {
	prometheusPort int
	prometheusPath string
	cluster        string
	awsRegion      string
}

var (
	consumerGroupLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_group_lag",
			Help: "The lag of consumer group break down by consumer_group, topic and partition",
		},
		[]string{"cluster", "consumer_group", "topic"},
	)

	mutexForUpdatingPromConsumerLagMetrics sync.Mutex

	ProvidedKafkaMonitorConsumerLagFlags KafkaMonitorConsumerLagFlags

	kafkaMonitorConsumerLagCmd = &cobra.Command{
		Use:   "consumer-lag",
		Short: "Expose prometheus metrics around kafka consumer group lags",
		RunE:  prometheusConsumerLagStatus,
		PreRun: func(cmd *cobra.Command, _ []string) {
			err := cmd.MarkPersistentFlagRequired("cluster")
			if err != nil {
				panic(err)
			}
		},
		Args: cobra.ExactArgs(0),
	}
)

// updateKafkaConsumerLagMetrics updates continuously prometheus metrics related to kafka topics and consumerGroup offset replication
func updateKafkaConsumerLagMetrics(cluster string, region string, cfgFile string, roleBasedAuth bool) {
	for {
		var totalLag int64
		consumerGroups, err := helper.GetConsumerGroups(cluster, region, cfgFile, []string{}, roleBasedAuth)
		if err != nil {
			log.Fatal("Unable to grab consumer groups of cluster :", cluster, err)
		}
		wg := sync.WaitGroup{}
		wg.Add(len(consumerGroups))
		for _, consumer := range consumerGroups {
			go func(consumer string) {
				lags, err := helper.GetConsumerGroupLag(cluster, region, cfgFile, consumer, []string{}, true, true)
				if err != nil {
					log.Fatal("Unable to grab consumer group lags of cluster :", cluster, err)
				}
				for topic, partitions := range lags {
					totalLag = 0
					for _, partitionLag := range partitions {
						totalLag += partitionLag
					}
					mutexForUpdatingPromMetrics.Lock()
					consumerGroupLag.WithLabelValues(cluster, consumer, topic).Set(float64(totalLag))
					mutexForUpdatingPromMetrics.Unlock()
				}
				wg.Done()
			}(consumer)
		}
		wg.Wait()
	}
}

func prometheusConsumerLagStatus(_ *cobra.Command, _ []string) error {
	prometheus.MustRegister(consumerGroupLag)
	// HTTP handler to expose Prometheus metrics
	http.Handle(ProvidedKafkaMonitorConsumerLagFlags.prometheusPath, promhttp.Handler())
	cluster := ProvidedKafkaMonitorConsumerLagFlags.cluster
	region := ProvidedKafkaMonitorConsumerLagFlags.awsRegion
	roleBasedAuth := ProvidedKafkaFlags.roleBasedAuth
	cfgFile := ProvidedKafkaFlags.cfgFile
	go updateKafkaConsumerLagMetrics(cluster, region, cfgFile, roleBasedAuth)
	// Start the HTTP server
	port := ProvidedKafkaMonitorConsumerLagFlags.prometheusPort
	cli.Printf("Server listening on :%d\n", port)
	err := http.ListenAndServe(":"+fmt.Sprint(port), nil) //#nosec G114
	if err != nil {
		cli.Println("Error starting server:", err)
		return err
	}
	return nil
}

func init() {
	kafkaMonitorConsumerLagCmd.PersistentFlags().IntVarP(&ProvidedKafkaMonitorConsumerLagFlags.prometheusPort, "port", "r", 4000, "port for prometheus server")
	kafkaMonitorConsumerLagCmd.PersistentFlags().StringVarP(&ProvidedKafkaMonitorConsumerLagFlags.cluster, "cluster", "c", "", "name of the kafka cluster")
	kafkaMonitorConsumerLagCmd.PersistentFlags().StringVarP(&ProvidedKafkaMonitorConsumerLagFlags.awsRegion, "awsregion", "R", "", "AWS region of the kafka cluster if any")
	kafkaMonitorConsumerLagCmd.PersistentFlags().StringVarP(&ProvidedKafkaMonitorConsumerLagFlags.prometheusPath, "path", "m", "/metrics", "metrics path for prometheus server")
}
