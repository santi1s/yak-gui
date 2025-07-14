package rollouts

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	providedStatusFlags RolloutsStatusFlags

	rolloutStatusCode = map[string]int{
		"Healthy":     0,
		"Progressing": 1,
		"Degraded":    2,
		"Paused":      3,
		"Error":       4,
	}
)

type RolloutsStatusFlags struct {
	rollout  string
	all      bool
	watch    bool
	timeout  time.Duration
	interval time.Duration
}

type statusRecord struct {
	Name   string
	Status string
}

type statusMap struct {
	Name        string
	Namespace   string
	Status      string
	Replicas    string
	Updated     string
	Ready       string
	Available   string
	Strategy    string
	CurrentStep string
	Revision    string
	Message     string
	Analysis    string
}

func getRolloutStatusData(namespace, rolloutName string) (map[string]statusMap, error) {
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %s", err)
	}

	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	var rollouts *unstructured.UnstructuredList
	if rolloutName != "" {
		rollout, err := dynamicClient.Resource(rolloutGVR).Namespace(namespace).Get(context.Background(), rolloutName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("rollout %s not found: %s", rolloutName, err)
		}
		rollouts = &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{*rollout},
		}
	} else {
		rollouts, err = dynamicClient.Resource(rolloutGVR).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list rollouts: %s", err)
		}
	}

	result := make(map[string]statusMap)
	for _, rollout := range rollouts.Items {
		status := buildRolloutStatus(&rollout)
		result[rollout.GetName()] = status
	}

	return result, nil
}

func buildRolloutStatus(rollout *unstructured.Unstructured) statusMap {
	name := rollout.GetName()
	namespace := rollout.GetNamespace()

	// Get spec replicas
	specReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "spec", "replicas")

	// Get status fields
	statusReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "replicas")
	updatedReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "updatedReplicas")
	readyReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "readyReplicas")
	availableReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "availableReplicas")

	// Get strategy
	strategy := "Unknown"
	if canarySpec, found, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "canary"); found && canarySpec != nil {
		strategy = "Canary"
	} else if blueGreenSpec, found, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "blueGreen"); found && blueGreenSpec != nil {
		strategy = "BlueGreen"
	}

	// Get current step
	currentStep := "N/A"
	if currentStepIndex, found, _ := unstructured.NestedInt64(rollout.Object, "status", "currentStepIndex"); found {
		currentStep = strconv.FormatInt(currentStepIndex+1, 10)
	}

	// Get phase/status
	phase, _, _ := unstructured.NestedString(rollout.Object, "status", "phase")
	if phase == "" {
		phase = "Unknown"
	}

	// Get message and enhance it with replica counts
	message, _, _ := unstructured.NestedString(rollout.Object, "status", "message")
	if message == "" {
		message = "<none>"
	} else {
		message = enhanceStatusMessage(rollout, message, specReplicas, statusReplicas, updatedReplicas, availableReplicas)
	}

	// Check for pod issues and append to message
	podIssues := checkPodIssues(rollout)
	if podIssues != "" {
		if message != "<none>" {
			message = fmt.Sprintf("%s - %s", message, podIssues)
		} else {
			message = podIssues
		}
	}

	// Get analysis run information
	analysis := getAnalysisStatus(rollout)

	// Get revision information
	revision := getRevisionInfo(rollout)

	return statusMap{
		Name:        name,
		Namespace:   namespace,
		Status:      phase,
		Replicas:    fmt.Sprintf("%d/%d", statusReplicas, specReplicas),
		Updated:     strconv.FormatInt(updatedReplicas, 10),
		Ready:       strconv.FormatInt(readyReplicas, 10),
		Available:   strconv.FormatInt(availableReplicas, 10),
		Strategy:    strategy,
		CurrentStep: currentStep,
		Revision:    revision,
		Message:     message,
		Analysis:    analysis,
	}
}

// enhanceStatusMessage adds specific replica counts to generic status messages
func enhanceStatusMessage(rollout *unstructured.Unstructured, message string, specReplicas, statusReplicas, updatedReplicas, availableReplicas int64) string {
	switch message {
	case "more replicas need to be updated":
		if specReplicas > 0 {
			remainingReplicas := specReplicas - updatedReplicas
			return fmt.Sprintf("%d more replicas need to be updated", remainingReplicas)
		}
	case "updated replicas are still becoming available":
		if updatedReplicas > availableReplicas {
			pendingReplicas := updatedReplicas - availableReplicas
			return fmt.Sprintf("%d updated replicas are still becoming available", pendingReplicas)
		}
	case "old replicas are pending termination":
		if statusReplicas > updatedReplicas {
			oldReplicas := statusReplicas - updatedReplicas
			return fmt.Sprintf("%d old replicas are pending termination", oldReplicas)
		}
	}
	return message
}

// getAnalysisStatus extracts analysis run information from the rollout status
func getAnalysisStatus(rollout *unstructured.Unstructured) string {
	// Check for canary analysis runs
	if canaryStatus, found, _ := unstructured.NestedMap(rollout.Object, "status", "canary"); found {
		if currentAnalysis, found, _ := unstructured.NestedMap(canaryStatus, "currentStepAnalysisRunStatus"); found {
			if status, found, _ := unstructured.NestedString(currentAnalysis, "status"); found {
				if name, nameFound, _ := unstructured.NestedString(currentAnalysis, "name"); nameFound {
					return fmt.Sprintf("Analysis: %s (%s)", name, status)
				}
				return fmt.Sprintf("Analysis: %s", status)
			}
		}

		// Check for background analysis
		if bgAnalysis, found, _ := unstructured.NestedMap(canaryStatus, "currentBackgroundAnalysisRunStatus"); found {
			if status, found, _ := unstructured.NestedString(bgAnalysis, "status"); found {
				if name, nameFound, _ := unstructured.NestedString(bgAnalysis, "name"); nameFound {
					return fmt.Sprintf("Analysis: %s (%s)", name, status)
				}
				return fmt.Sprintf("Analysis: %s", status)
			}
		}
	}

	// Check for blue-green post-promotion analysis
	if blueGreenStatus, found, _ := unstructured.NestedMap(rollout.Object, "status", "blueGreen"); found {
		if postPromotionAnalysis, found, _ := unstructured.NestedMap(blueGreenStatus, "postPromotionAnalysisRunStatus"); found {
			if status, found, _ := unstructured.NestedString(postPromotionAnalysis, "status"); found {
				if name, nameFound, _ := unstructured.NestedString(postPromotionAnalysis, "name"); nameFound {
					return fmt.Sprintf("Analysis: %s (%s)", name, status)
				}
				return fmt.Sprintf("Analysis: %s", status)
			}
		}
	}

	return ""
}

// getRevisionInfo extracts current revision information from the rollout
func getRevisionInfo(rollout *unstructured.Unstructured) string {
	// Get the current revision from rollout annotations (this is how argo rollouts tracks revisions)
	annotations := rollout.GetAnnotations()
	if annotations != nil {
		if revisionStr, exists := annotations["rollout.argoproj.io/revision"]; exists && revisionStr != "" {
			return fmt.Sprintf("revision:%s", revisionStr)
		}
	}

	// Fallback to generation if no revision annotation
	generation := rollout.GetGeneration()
	return fmt.Sprintf("generation:%d", generation)
}

// checkPodIssues examines rollout conditions and status to detect pod issues
func checkPodIssues(rollout *unstructured.Unstructured) string {
	// Check rollout conditions for ReplicaFailure
	conditions, found, _ := unstructured.NestedSlice(rollout.Object, "status", "conditions")
	if found {
		for _, condition := range conditions {
			condMap, ok := condition.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _ := condMap["type"].(string)
			status, _ := condMap["status"].(string)
			reason, _ := condMap["reason"].(string)
			message, _ := condMap["message"].(string)

			// Check for ReplicaFailure condition
			if condType == "ReplicaFailure" && status == "True" {
				if reason != "" {
					return fmt.Sprintf("ReplicaFailure: %s", reason)
				}
				if message != "" {
					return fmt.Sprintf("ReplicaFailure: %s", message)
				}
				return "ReplicaFailure"
			}

			// Check for Progressing condition with failure reason
			if condType == "Progressing" && status == "False" {
				if reason == "ProgressDeadlineExceeded" {
					return "ProgressDeadlineExceeded"
				}
				if reason != "" {
					return fmt.Sprintf("Progress failed: %s", reason)
				}
				if message != "" {
					return fmt.Sprintf("Progress failed: %s", message)
				}
			}
		}
	}

	// Try to get additional information from current pod hash status
	currentPodHash, _, _ := unstructured.NestedString(rollout.Object, "status", "currentPodHash")
	if currentPodHash != "" {
		// If we have updated replicas but ready replicas are less, check for pod issues
		updatedReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "updatedReplicas")
		readyReplicas, _, _ := unstructured.NestedInt64(rollout.Object, "status", "readyReplicas")

		if updatedReplicas > 0 && readyReplicas < updatedReplicas {
			// There are updated replicas that are not ready - this suggests pod issues
			// Check if this looks like an image pull issue based on the pattern
			phase, _, _ := unstructured.NestedString(rollout.Object, "status", "phase")
			if phase == "Progressing" && readyReplicas == 0 {
				return "Pods not becoming ready (possible ImagePullBackOff)"
			}
		}
	}

	return ""
}

func status(cmd *cobra.Command, args []string) error {
	namespace := resolveNamespace()
	if providedStatusFlags.all {
		namespace = ""
	}

	// Handle watch mode for single rollout
	if providedStatusFlags.watch && providedStatusFlags.rollout != "" {
		watchOptions := WatchStatusOptions{
			Watch:    providedStatusFlags.watch,
			Timeout:  providedStatusFlags.timeout,
			Interval: providedStatusFlags.interval,
		}
		return watchRolloutStatus(namespace, providedStatusFlags.rollout, watchOptions)
	}

	// Handle non-watch mode or multiple rollouts
	result, err := getRolloutStatusData(namespace, providedStatusFlags.rollout)
	if err != nil {
		return err
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(result)
	} else {
		formatOutputStatus(result)
	}
	return nil
}

func formatOutputStatus(status map[string]statusMap) {
	const nameLabel = "NAME"
	const namespaceLabel = "NAMESPACE"
	const statusLabel = "STATUS"
	const replicasLabel = "REPLICAS"
	const updatedLabel = "UPDATED"
	const readyLabel = "READY"
	const availableLabel = "AVAILABLE"
	const strategyLabel = "STRATEGY"
	const stepLabel = "STEP"
	const revisionLabel = "REVISION"
	const analysisLabel = "ANALYSIS"
	const messageLabel = "MESSAGE"

	var nameWidth = len(nameLabel)
	var namespaceWidth = len(namespaceLabel)
	var statusWidth = len(statusLabel)
	var replicasWidth = len(replicasLabel)
	var updatedWidth = len(updatedLabel)
	var readyWidth = len(readyLabel)
	var availableWidth = len(availableLabel)
	var strategyWidth = len(strategyLabel)
	var stepWidth = len(stepLabel)
	var revisionWidth = len(revisionLabel)
	var analysisWidth = len(analysisLabel)

	names := make([]string, 0, len(status))
	for name := range status {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		s := status[name]
		nameWidth = max(nameWidth, len(s.Name))
		namespaceWidth = max(namespaceWidth, len(s.Namespace))
		statusWidth = max(statusWidth, len(s.Status))
		replicasWidth = max(replicasWidth, len(s.Replicas))
		updatedWidth = max(updatedWidth, len(s.Updated))
		readyWidth = max(readyWidth, len(s.Ready))
		availableWidth = max(availableWidth, len(s.Available))
		strategyWidth = max(strategyWidth, len(s.Strategy))
		stepWidth = max(stepWidth, len(s.CurrentStep))
		revisionWidth = max(revisionWidth, len(s.Revision))
		analysisWidth = max(analysisWidth, len(s.Analysis))
	}

	nameWidth++
	namespaceWidth++
	statusWidth++
	replicasWidth++
	updatedWidth++
	readyWidth++
	availableWidth++
	strategyWidth++
	stepWidth++
	revisionWidth++
	analysisWidth++

	cli.Printf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %s\n",
		nameWidth, nameLabel,
		namespaceWidth, namespaceLabel,
		statusWidth, statusLabel,
		replicasWidth, replicasLabel,
		updatedWidth, updatedLabel,
		readyWidth, readyLabel,
		availableWidth, availableLabel,
		strategyWidth, strategyLabel,
		stepWidth, stepLabel,
		revisionWidth, revisionLabel,
		analysisWidth, analysisLabel,
		messageLabel)

	for _, name := range names {
		s := status[name]
		cli.Printf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s %s\n",
			nameWidth, s.Name,
			namespaceWidth, s.Namespace,
			statusWidth, s.Status,
			replicasWidth, s.Replicas,
			updatedWidth, s.Updated,
			readyWidth, s.Ready,
			availableWidth, s.Available,
			strategyWidth, s.Strategy,
			stepWidth, s.CurrentStep,
			revisionWidth, s.Revision,
			analysisWidth, s.Analysis,
			s.Message)
	}
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of one or all rollouts",
	Example: `yak rollouts status -r my-rollout
yak rollouts status --all
yak rollouts status -r my-rollout --watch=false
yak rollouts status -r my-rollout --timeout 60s`,
	RunE: status,
}

func init() {
	statusCmd.Flags().StringVarP(&providedStatusFlags.rollout, "rollout", "r", "", "Rollout name")
	statusCmd.Flags().BoolVar(&providedStatusFlags.all, "all", false, "Show rollouts from all namespaces")
	statusCmd.Flags().BoolVarP(&providedStatusFlags.watch, "watch", "w", true, "Watch the status of the rollout until it's done")
	statusCmd.Flags().DurationVarP(&providedStatusFlags.timeout, "timeout", "t", 0, "The length of time to watch before giving up (e.g. 1s, 2m, 3h). Zero means wait forever")
	statusCmd.Flags().DurationVar(&providedStatusFlags.interval, "interval", 2*time.Second, "The polling interval when watching")
}
