package argocd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/santi1s/yak/cli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/spf13/cobra"
)

var (
	orphanCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "argocd_orphan_resources_count",
			Help: "Count of orphan resources aggregated by appName",
		},
		[]string{"appName"},
	)

	statusSyncCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "argocd_status_sync_count",
			Help: "Application sync status aggregated by appName (0: Synced, 1: OutOfSync, 2: Unknown))",
		},
		[]string{"appName"},
	)

	statusSuspendedCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "argocd_status_suspended_count",
			Help: "Applications suspended status aggregated by appName (0: unsuspended, 1: suspended))",
		},
		[]string{"appName"},
	)

	statusHealthCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "argocd_status_health_count",
			Help: "Applications health status aggregated by appName (0: Healthy, 1: Degraded, 2: Progressing, 3: Suspended, 4: Missing, 5: Unknown))",
		},
		[]string{"appName"},
	)

	mutexForUpdatingPromMetrics sync.Mutex

	providedMonitoringFlags ArgoCDMonitorFlags

	monitorArgoCDCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Collect all metrics related to the monitoring of argocd and expose them in a prometheus format",
		RunE:  monitorArgoCDResources,
	}
)

type ArgoCDMonitorFlags struct {
	prometheusEnable bool
	prometheusPath   string
	prometheusPort   int
	username         string
	expose           bool
}

type orphanedResourcesMonitorRecord struct {
	count   int
	appName string
}

func constructOrphanedResourcesMonitorRecord(orphanedResourcesByApp map[string][]argocdhelper.AppResource, orphanedNonNamespacedResources []argocdhelper.AppResource) ([]orphanedResourcesMonitorRecord, error) {
	var result []orphanedResourcesMonitorRecord

	for appName, orphanResources := range orphanedResourcesByApp {
		for _, ors := range orphanResources {
			log.Debug(fmt.Sprintf("appName: %s, namespace: %s, kind: %s, name: %s, group: %s \n", appName, ors.Namespace, ors.Kind, ors.Name, ors.Group))
		}
		result = append(result, orphanedResourcesMonitorRecord{appName: appName, count: len(orphanResources)})
	}
	for _, ors := range orphanedNonNamespacedResources {
		log.Debug(fmt.Sprintf("appName: --, namespace: %s, kind: %s, name: %s, group: %s \n", ors.Namespace, ors.Kind, ors.Name, ors.Group))
	}
	result = append(result, orphanedResourcesMonitorRecord{appName: "--", count: len(orphanedNonNamespacedResources)})
	return result, nil
}

func getStatusRecord(project *v1alpha1.AppProject, apps *v1alpha1.ApplicationList) (map[string][]statusRecord, error) {
	result := make(map[string][]statusRecord)

	statusData, err := buildStatusData(project, apps)
	if err != nil {
		return nil, fmt.Errorf("unable to get status data by app: %s", err)
	}
	for appName, status := range statusData {
		result["sync"] = append(result["sync"], statusRecord{AppName: appName, Count: syncStatusCode[string(status.Sync)]})
		result["health"] = append(result["health"], statusRecord{AppName: appName, Count: healthStatusCode[string(status.Health)]})
		suspendedMetric := 0
		if status.Suspended {
			suspendedMetric = 1
		}
		result["suspended"] = append(result["suspended"], statusRecord{AppName: appName, Count: suspendedMetric})
	}

	return result, nil
}

func updateArgoCDMetrics(loginParams *argocdhelper.LoginParams, projectName string, serverMode bool) {
	for {
		apiclient, err := argocdhelper.ArgocdLogin(loginParams)
		if err != nil {
			log.Fatal("Unable to establish connection to argocd: ", err)
		}

		// orphaned Namespaced Resources
		orphanedResourcesByApp, err := argocdhelper.OrphanedResourcesArgoCD(apiclient.AppClient, projectName)

		if err != nil {
			log.Fatal("Unable to get orphaned resources by app", err)
		}

		// Get orphaned non-namespaced Resources
		remainingOrphanedResources, err := argocdhelper.GetOrphanedNonNamespacedResources(apiclient.AppClient, projectName)

		if err != nil {
			log.Fatal("Unable to get non-namespaced orphan resources", err)
		}

		orphans, err := constructOrphanedResourcesMonitorRecord(orphanedResourcesByApp, remainingOrphanedResources)
		if err != nil {
			log.Fatal("Unable to construct orphan resources monitor records", err)
		}

		myProject, err := argocdhelper.GetArgoCDProject(apiclient.ProjectClient, projectName)
		if err != nil {
			log.Fatalf("Project getting list failed: %v", err)
		}

		myApps, err := apiclient.AppClient.List(context.Background(), &application.ApplicationQuery{Project: []string{myProject.Name}})
		if err != nil {
			log.Fatalf("Application list failed: %v", err)
		}
		// Get the data & transform it to prometheus format
		status, err := getStatusRecord(myProject, myApps)
		if err != nil {
			log.Fatal("Unable to construct status resources monitor records", err)
		}

		for _, orphan := range orphans {
			mutexForUpdatingPromMetrics.Lock()
			orphanCount.WithLabelValues(orphan.appName).Set(float64(orphan.count))
			mutexForUpdatingPromMetrics.Unlock()
		}

		for _, sync := range status["sync"] {
			mutexForUpdatingPromMetrics.Lock()
			statusSyncCount.WithLabelValues(sync.AppName).Set(float64(sync.Count))
			mutexForUpdatingPromMetrics.Unlock()
		}
		for _, suspended := range status["suspended"] {
			mutexForUpdatingPromMetrics.Lock()
			statusSuspendedCount.WithLabelValues(suspended.AppName).Set(float64(suspended.Count))
			mutexForUpdatingPromMetrics.Unlock()
		}

		for _, health := range status["health"] {
			mutexForUpdatingPromMetrics.Lock()
			statusHealthCount.WithLabelValues(health.AppName).Set(float64(health.Count))
			mutexForUpdatingPromMetrics.Unlock()
		}

		if !serverMode {
			// Gather the metrics
			metrics, err := prometheus.DefaultGatherer.Gather()
			if err != nil {
				log.Fatal("Unable to gather prometheus metrics", err)
			}

			// Format the metrics as plain text
			out := &bytes.Buffer{}
			for _, mf := range metrics {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					log.Fatal("Unable to format prometheus metrics to text", err)
				}
			}

			fmt.Print(out.String())
			break
		}
		// Sleep for a period of 5 min before updating metrics again
		sleepDuration := 5 * time.Minute
		time.Sleep(sleepDuration)
	}
}

func monitorArgoCDResources(_ *cobra.Command, _ []string) error {
	prometheus.MustRegister(orphanCount)
	prometheus.MustRegister(statusSyncCount)
	prometheus.MustRegister(statusSuspendedCount)
	prometheus.MustRegister(statusHealthCount)
	// HTTP handler to expose Prometheus metrics
	http.Handle(providedMonitoringFlags.prometheusPath, promhttp.Handler())

	// Set params from provided flags
	LoginParams := argocdhelper.LoginParams{
		ArgocdServer:   providedFlags.addr,
		ArgocdUsername: providedMonitoringFlags.username,
		ArgocdPassword: os.Getenv("MONITORING_PASSWORD"),
	}

	if providedMonitoringFlags.expose {
		go updateArgoCDMetrics(&LoginParams, providedFlags.project, providedMonitoringFlags.expose)

		// Start the HTTP server
		port := providedMonitoringFlags.prometheusPort
		cli.Printf("Server listening on :%d\n", port)
		err := http.ListenAndServe(":"+fmt.Sprint(port), nil) //#nosec G114
		if err != nil {
			cli.Println("Error starting server:", err)
			return err
		}
	} else {
		updateArgoCDMetrics(&LoginParams, providedFlags.project, providedMonitoringFlags.expose)
	}

	return nil
}

func init() {
	monitorArgoCDCmd.Flags().StringVarP(&providedMonitoringFlags.username, "username", "u", "monitoring", "Argocd username")
	monitorArgoCDCmd.Flags().IntVarP(&providedMonitoringFlags.prometheusPort, "prom-port", "p", 4000, "prometheus port")
	monitorArgoCDCmd.Flags().StringVarP(&providedMonitoringFlags.prometheusPath, "prom-path", "m", "/metrics", "prometheus path")
	monitorArgoCDCmd.Flags().BoolVar(&providedMonitoringFlags.expose, "expose", false, "expose metrics")
}
