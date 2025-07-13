package rollouts

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var providedHistoryFlags RolloutsHistoryFlags

type RolloutsHistoryFlags struct {
	rollout  string
	revision int64
}

type revisionInfo struct {
	Revision   int64
	StartedAt  string
	FinishedAt string
	Duration   string
	Phase      string
	Message    string
}

func history(cmd *cobra.Command, args []string) error {
	rolloutName := providedHistoryFlags.rollout
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	// Get the rollout
	rollout, err := dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Get(context.Background(), rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found in namespace %s: %s", rolloutName, resolveNamespace(), err)
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(rollout)
	} else {
		formatRolloutHistory(rollout, dynamicClient)
	}
	return nil
}

func formatRolloutHistory(rollout *unstructured.Unstructured, dynamicClient dynamic.Interface) {
	name := rollout.GetName()
	cli.Printf("Rollout History: %s\n", name)
	cli.Printf("---\n")

	// Get revision history from multiple sources
	revisions := []revisionInfo{}

	// 1. Check rollout status for current/stable revisions
	currentRevisions := getCurrentRevisions(rollout)
	revisions = append(revisions, currentRevisions...)

	// 2. Get ReplicaSets owned by this rollout
	rsRevisions, err := getReplicaSetRevisions(rollout, dynamicClient)
	if err == nil {
		revisions = append(revisions, rsRevisions...)
	}

	// Remove duplicates and sort
	revisions = deduplicateRevisions(revisions)
	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].Revision > revisions[j].Revision
	})

	if len(revisions) == 0 {
		cli.Printf("No revision history found\n")
		return
	}

	// Display specific revision if requested
	if providedHistoryFlags.revision > 0 {
		for _, rev := range revisions {
			if rev.Revision == providedHistoryFlags.revision {
				formatRevisionDetails(rev)
				return
			}
		}
		cli.Printf("Revision %d not found\n", providedHistoryFlags.revision)
		return
	}

	// Display all revisions
	const revisionLabel = "REVISION"
	const startedLabel = "STARTED"
	const statusLabel = "STATUS"
	const hashLabel = "POD-HASH"
	const ageLabel = "AGE"

	var revisionWidth = len(revisionLabel)
	var startedWidth = len(startedLabel)
	var statusWidth = len(statusLabel)
	var hashWidth = len(hashLabel)
	var ageWidth = len(ageLabel)

	for _, rev := range revisions {
		revisionWidth = max(revisionWidth, len(strconv.FormatInt(rev.Revision, 10)))
		startedWidth = max(startedWidth, len(rev.StartedAt))
		statusWidth = max(statusWidth, len(rev.Phase))
		hashWidth = max(hashWidth, len(rev.Message))
		ageWidth = max(ageWidth, len(rev.Duration))
	}

	revisionWidth++
	startedWidth++
	statusWidth++
	hashWidth++
	ageWidth++

	cli.Printf("%-*s %-*s %-*s %-*s %s\n",
		revisionWidth, revisionLabel,
		startedWidth, startedLabel,
		statusWidth, statusLabel,
		hashWidth, hashLabel,
		ageLabel)

	for _, rev := range revisions {
		cli.Printf("%-*d %-*s %-*s %-*s %s\n",
			revisionWidth, rev.Revision,
			startedWidth, rev.StartedAt,
			statusWidth, rev.Phase,
			hashWidth, rev.Message,
			rev.Duration)
	}
}

func getCurrentRevisions(rollout *unstructured.Unstructured) []revisionInfo {
	var revisions []revisionInfo

	if rollout == nil {
		return revisions
	}

	// Try to get revision from rollout annotations
	annotations := rollout.GetAnnotations()
	if annotations != nil {
		if revStr, exists := annotations["rollout.argoproj.io/revision"]; exists && revStr != "" {
			if rev, err := strconv.ParseInt(revStr, 10, 64); err == nil {
				revisions = append(revisions, revisionInfo{Revision: rev})
				return revisions
			}
		}
	}

	// Fallback to generation if no revision annotation
	generation := rollout.GetGeneration()
	revisions = append(revisions, revisionInfo{Revision: generation})

	return revisions
}

func getReplicaSetRevisions(rollout *unstructured.Unstructured, dynamicClient dynamic.Interface) ([]revisionInfo, error) {
	var revisions []revisionInfo

	// Get ReplicaSets owned by this rollout
	rsGVR := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "replicasets",
	}

	// Get the selector from the rollout
	selector, found, err := unstructured.NestedStringMap(rollout.Object, "spec", "selector", "matchLabels")
	if err != nil || !found {
		return revisions, fmt.Errorf("unable to get selector from rollout")
	}

	// Convert map to label selector string
	var selectorParts []string
	for key, value := range selector {
		selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", key, value))
	}
	labelSelector := strings.Join(selectorParts, ",")

	rsList, err := dynamicClient.Resource(rsGVR).Namespace(rollout.GetNamespace()).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return revisions, err
	}

	for _, rs := range rsList.Items {
		revision := extractRolloutRevisionInfo(&rs)
		if revision.Revision > 0 {
			revisions = append(revisions, revision)
		}
	}

	return revisions, nil
}

func extractRolloutRevisionInfo(rs *unstructured.Unstructured) revisionInfo {
	// Get the actual revision number from annotations
	annotations := rs.GetAnnotations()
	var revision int64 = 0

	if annotations != nil {
		// Check for Argo Rollouts revision annotation
		if revStr, exists := annotations["rollout.argoproj.io/revision"]; exists {
			if rev, err := strconv.ParseInt(revStr, 10, 64); err == nil {
				revision = rev
			}
		}
		// Fallback to deployment revision annotation
		if revision == 0 {
			if revStr, exists := annotations["deployment.kubernetes.io/revision"]; exists {
				if rev, err := strconv.ParseInt(revStr, 10, 64); err == nil {
					revision = rev
				}
			}
		}
	}

	// If no revision found in annotations, return empty
	if revision == 0 {
		return revisionInfo{}
	}

	// For Argo Rollouts, look for rollouts-pod-template-hash label
	labels := rs.GetLabels()
	podHash := ""
	if labels != nil {
		podHash = labels["rollouts-pod-template-hash"]
	}

	// Get creation time
	creationTime := rs.GetCreationTimestamp()
	startedAt := "Unknown"
	if !creationTime.IsZero() {
		startedAt = creationTime.Format("2006-01-02 15:04:05")
	}

	// Get replica counts to determine status
	specReplicas, _, _ := unstructured.NestedInt64(rs.Object, "spec", "replicas")
	statusReplicas, _, _ := unstructured.NestedInt64(rs.Object, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(rs.Object, "status", "readyReplicas")

	phase := "Unknown"
	finishedAt := "Active"
	duration := helper.GetAge(creationTime.Time)

	if specReplicas == 0 {
		phase = "Scaled Down"
		finishedAt = "Completed"
	} else if readyReplicas == specReplicas {
		phase = "Ready"
	} else if statusReplicas > 0 {
		phase = "Progressing"
	} else {
		phase = "Pending"
	}

	return revisionInfo{
		Revision:   revision,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Duration:   duration,
		Phase:      phase,
		Message:    podHash,
	}
}

func deduplicateRevisions(revisions []revisionInfo) []revisionInfo {
	seen := make(map[int64]revisionInfo)

	for _, rev := range revisions {
		// Use revision number as the primary key for deduplication
		if existing, exists := seen[rev.Revision]; exists {
			// Prefer "current" over other phases
			if rev.Phase == "current" && existing.Phase != "current" {
				seen[rev.Revision] = rev
			} else if existing.Phase == "current" && rev.Phase != "current" {
				// Keep the existing "current" one
				continue
			} else if rev.Phase != "" && existing.Phase == "" {
				// Prefer non-empty phase over empty phase
				seen[rev.Revision] = rev
			}
		} else {
			seen[rev.Revision] = rev
		}
	}

	// Convert map back to slice
	var result []revisionInfo
	for _, rev := range seen {
		result = append(result, rev)
	}

	// Sort by revision number
	sort.Slice(result, func(i, j int) bool {
		return result[i].Revision > result[j].Revision
	})

	return result
}

func extractRevisionInfo(rs *unstructured.Unstructured) revisionInfo {
	// Get revision from deployment-revision annotation
	annotations := rs.GetAnnotations()
	if annotations == nil {
		return revisionInfo{}
	}

	revisionStr, exists := annotations["deployment.kubernetes.io/revision"]
	if !exists {
		return revisionInfo{}
	}

	revision, err := strconv.ParseInt(revisionStr, 10, 64)
	if err != nil {
		return revisionInfo{}
	}

	// Get creation timestamp
	creationTime := rs.GetCreationTimestamp()
	startedAt := "Unknown"
	if !creationTime.IsZero() {
		startedAt = creationTime.Format("2006-01-02 15:04:05")
	}

	// For finished time and duration, we'd need to track when the ReplicaSet was scaled down
	// This is a simplified implementation
	finishedAt := "Active"
	duration := "Unknown"

	// Check if ReplicaSet is still active (has replicas)
	replicas, _, _ := unstructured.NestedInt64(rs.Object, "status", "replicas")
	if replicas == 0 {
		finishedAt = "Completed"
		if !creationTime.IsZero() {
			// Estimate duration (this is simplified)
			duration = time.Since(creationTime.Time).Truncate(time.Second).String()
		}
	}

	phase := "Unknown"
	readyReplicas, _, _ := unstructured.NestedInt64(rs.Object, "status", "readyReplicas")
	availableReplicas, _, _ := unstructured.NestedInt64(rs.Object, "status", "availableReplicas")

	if replicas > 0 {
		if readyReplicas == replicas && availableReplicas == replicas {
			phase = "Active"
		} else {
			phase = "Progressing"
		}
	} else {
		phase = "Completed"
	}

	message := "N/A"
	if conditions, found, _ := unstructured.NestedSlice(rs.Object, "status", "conditions"); found && len(conditions) > 0 {
		// Get the latest condition message
		if condMap, ok := conditions[len(conditions)-1].(map[string]interface{}); ok {
			if msg, ok := condMap["message"].(string); ok && msg != "" {
				message = msg
			}
		}
	}

	return revisionInfo{
		Revision:   revision,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Duration:   duration,
		Phase:      phase,
		Message:    message,
	}
}

func formatRevisionDetails(rev revisionInfo) {
	cli.Printf("Revision: %d\n", rev.Revision)
	cli.Printf("Started At: %s\n", rev.StartedAt)
	cli.Printf("Finished At: %s\n", rev.FinishedAt)
	cli.Printf("Duration: %s\n", rev.Duration)
	cli.Printf("Phase: %s\n", rev.Phase)
	cli.Printf("Message: %s\n", rev.Message)
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show rollout revision history",
	Example: `yak rollouts history -r my-rollout
yak rollouts history -r my-rollout --revision 3`,
	RunE: history,
}

func init() {
	historyCmd.Flags().StringVarP(&providedHistoryFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	historyCmd.Flags().Int64Var(&providedHistoryFlags.revision, "revision", 0, "Show details for specific revision")
	_ = historyCmd.MarkFlagRequired("rollout")
}
