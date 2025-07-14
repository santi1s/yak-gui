package argocd

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/santi1s/yak/cli"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

var (
	providedStatusFlags ArgoCDStatusFlags

	syncStatusCode = map[string]int{
		"Synced":    0,
		"OutOfSync": 1,
		"Unknown":   2,
	}

	healthStatusCode = map[string]int{
		// 100% healthy
		"Healthy": 0,
		// Not healthy but still have a chance to reach healthy state
		"Progressing": 1,
		// Status indicates failure or could not reach healthy state within some timeout.
		"Degraded": 2,
		// Assigned to resources that are suspended or paused (e.g suspended CronJob)
		"Suspended": 3,
		// Resource is missing in the cluster.
		"Missing": 4,
		// health assessment failed and actual health status is unknown
		"Unknown": 5,
	}
)

type ArgoCDStatusFlags struct {
	application      string
	prometheusEnable bool
	prometheusPath   string
	prometheusPort   int
	suspendedOnly    bool
	outOfSyncOnly    bool
}

type statusRecord struct {
	AppName string
	Count   int
}

type statusMap struct {
	AppName    string
	Health     string
	Conditions []string
	Sync       string
	Suspended  bool
	SyncLoop   string
}

func getAppStatusData(projectName, appName string) (map[string]statusMap, error) {
	// Authentication
	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	myProject, err := argocdhelper.GetArgoCDProject(apiclient.ProjectClient, projectName)
	if err != nil {
		return nil, fmt.Errorf("project getting list failed: %s", err)
	}

	// Get the list of applications from argocd-apps namespace
	var argocdApps *v1alpha1.ApplicationList
	if appName != "" {
		// If an application is provided, we only get the status of that application
		argocdApps, err = apiclient.AppClient.List(context.Background(), &application.ApplicationQuery{Project: []string{myProject.Name}, Name: &appName})
		if err != nil {
			return nil, fmt.Errorf("application %s list failed: %s", appName, err)
		}
	} else {
		argocdApps, err = apiclient.AppClient.List(context.Background(), &application.ApplicationQuery{Project: []string{myProject.Name}})
		if err != nil {
			return nil, fmt.Errorf("application list failed: %s", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("application list failed: %s", err)
	}

	// Get the metrics from monitor cmds
	status, err := buildStatusData(myProject, argocdApps)
	if err != nil {
		return nil, fmt.Errorf("an error occured on getting status of applications: %s", err)
	}

	return status, err
}

func buildStatusData(project *v1alpha1.AppProject, apps *v1alpha1.ApplicationList) (map[string]statusMap, error) {
	result := make(map[string]statusMap)

	for _, item := range apps.Items {
		app := item
		conditions := []string{}
		for _, condition := range app.Status.Conditions {
			conditions = append(conditions, condition.Type)
		}
		result[string(app.Name)] = statusMap{
			AppName:    app.Name,
			Health:     string(app.Status.Health.Status),
			Conditions: conditions,
			Sync:       string(app.Status.Sync.Status),
			Suspended:  isAppSuspended(project, &app),
			SyncLoop:   detectSyncLoop(&app),
		}
	}

	return result, nil
}

func isAppSuspended(project *v1alpha1.AppProject, app *v1alpha1.Application) bool {
	windows := project.Spec.SyncWindows.Matches(app)
	return windows != nil
}

func detectSyncLoop(app *v1alpha1.Application) string {
	// Check if app has sync history
	if len(app.Status.History) < 3 {
		return "No"
	}

	// Analyze recent sync history (last 10 deployments or all if less)
	historyCount := len(app.Status.History)
	recentCount := 10
	if historyCount < recentCount {
		recentCount = historyCount
	}

	recentHistory := app.Status.History[historyCount-recentCount:]

	// Check for rapid sync patterns
	var rapidSyncs int
	var recentSyncs int
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	fifteenMinutesAgo := now.Add(-15 * time.Minute)

	for _, deployment := range recentHistory {
		if !deployment.DeployedAt.IsZero() {
			deployTime := deployment.DeployedAt.Time

			// Count syncs in last hour
			if deployTime.After(oneHourAgo) {
				recentSyncs++
			}

			// Count syncs in last 15 minutes
			if deployTime.After(fifteenMinutesAgo) {
				rapidSyncs++
			}
		}
	}

	// Check current operation state
	hasRunningOperation := app.Status.OperationState != nil &&
		app.Status.OperationState.Phase == "Running"

	// Sync loop detection logic
	if rapidSyncs >= 3 {
		return "Critical" // 3+ syncs in 15 minutes
	} else if recentSyncs >= 6 {
		return "Warning" // 6+ syncs in 1 hour
	} else if hasRunningOperation && recentSyncs >= 3 {
		return "Possible" // Currently syncing with recent activity
	}

	// Check for repeated sync attempts without success (simplified logic)
	// In a real implementation, you'd check operation status history
	// For now, just check if we have many recent syncs but still OutOfSync
	if recentSyncs >= 4 && app.Status.Sync.Status == "OutOfSync" {
		return "Failed" // Many syncs but still out of sync
	}

	return "No"
}

func status(cmd *cobra.Command, args []string) error {
	result, err := getAppStatusData(providedFlags.project, providedStatusFlags.application)
	if err != nil {
		return err
	}

	if providedStatusFlags.suspendedOnly {
		filteredResult := make(map[string]statusMap)
		for appName, status := range result {
			if status.Suspended {
				filteredResult[appName] = status
			}
		}
		result = filteredResult
	}

	if providedStatusFlags.outOfSyncOnly {
		filteredResult := make(map[string]statusMap)
		for appName, status := range result {
			if status.Sync == "OutOfSync" {
				filteredResult[appName] = status
			}
		}
		result = filteredResult
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(result)
	} else {
		formatOutputStatus(result)
	}
	return nil
}

func formatOutputStatus(status map[string]statusMap) {
	// Define header labels
	const applicationLabel = "APPLICATION"
	const syncStatusLabel = "SYNC STATUS"
	const suspendedStatus = "SUSPENDED"
	const syncLoopLabel = "SYNC LOOP"
	const healthLabel = "HEALTH"
	const conditionsLabel = "CONDITIONS"
	// Init with minimum width
	var nameWidth = len(applicationLabel)
	var statusWidth = len(syncStatusLabel)
	var suspendedWidth = 5
	if len(suspendedStatus) > 5 {
		suspendedWidth = len(suspendedStatus)
	}
	var syncLoopWidth = len(syncLoopLabel)
	var healthWidth = len(healthLabel)
	apps := make([]string, 0, len(status))
	for app := range status {
		apps = append(apps, app)
	}
	sort.Strings(apps)

	// Increase width according to longest value
	for appName, status := range status {
		nameWidth = max(nameWidth, len(appName))
		statusWidth = max(statusWidth, len(status.Sync))
		syncLoopWidth = max(syncLoopWidth, len(status.SyncLoop))
		healthWidth = max(healthWidth, len(status.Health))
	}
	// Add more space
	nameWidth++
	statusWidth++
	suspendedWidth++
	syncLoopWidth++
	healthWidth++
	// Print
	cli.Printf("%-*s %-*s %-*s %-*s %-*s %s\n", nameWidth, applicationLabel, statusWidth, syncStatusLabel, suspendedWidth, suspendedStatus, syncLoopWidth, syncLoopLabel, healthWidth, healthLabel, conditionsLabel)

	for _, appName := range apps {
		conditionDisplay := strings.Join(status[appName].Conditions, ",")
		if conditionDisplay == "" {
			conditionDisplay = "<none>"
		}

		cli.Printf("%-*s %-*s %-*s %-*s %-*s %s\n", nameWidth, appName, statusWidth, status[appName].Sync, suspendedWidth, strconv.FormatBool(status[appName].Suspended), syncLoopWidth, status[appName].SyncLoop, healthWidth, status[appName].Health, conditionDisplay)
	}
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status for on or all the application (suspend & sync)",
	Example: `yak argocd status -a my-app
yak argocd status --suspended-only
yak argocd status --out-of-sync-only`,
	RunE: status,
}

func init() {
	statusCmd.Flags().StringVarP(&providedStatusFlags.application, "application", "a", "", "ArgoCD application name")
	statusCmd.Flags().BoolVarP(&providedStatusFlags.suspendedOnly, "suspended-only", "u", false, "Show only suspended applications")
	statusCmd.Flags().BoolVarP(&providedStatusFlags.outOfSyncOnly, "out-of-sync-only", "o", false, "Show only out-of-sync applications")
}
