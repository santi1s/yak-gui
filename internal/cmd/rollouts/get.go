package rollouts

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/helper"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var providedGetFlags RolloutsGetFlags

type RolloutsGetFlags struct {
	rollout string
}

func get(cmd *cobra.Command, args []string) error {
	rolloutName := providedGetFlags.rollout
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

	rollout, err := dynamicClient.Resource(rolloutGVR).Namespace(resolveNamespace()).Get(context.Background(), rolloutName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("rollout %s not found in namespace %s: %s", rolloutName, resolveNamespace(), err)
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(rollout)
	} else {
		err := formatRolloutDetails(rollout)
		if err != nil {
			return err
		}
	}
	return nil
}

func formatRolloutDetails(rollout *unstructured.Unstructured) error {
	// Format strategy details
	strategy := formatStrategy(rollout)

	// Format revision details
	revision := formatRevision(rollout)

	// Format conditions
	conditions := formatConditions(rollout)

	// Format analysis runs and get the raw data for selection
	analysisRuns := formatAnalysisRuns(rollout)

	// Create table
	table := tablewriter.NewWriter(os.Stdout)

	// Configure table
	table.SetHeader([]string{"STRATEGY", "REVISION", "CONDITIONS", "ANALYSIS RUNS"})
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator(" ")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.SetAutoWrapText(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColMinWidth(0, 30) // STRATEGY column
	table.SetColMinWidth(1, 25) // REVISION column
	table.SetColMinWidth(2, 40) // CONDITIONS column
	table.SetColMinWidth(3, 40) // ANALYSIS RUNS column

	// Add data row
	table.Append([]string{strategy, revision, conditions, analysisRuns})

	// Render table
	table.Render()

	// If there are analysis runs, offer interactive selection
	if analysisRuns != "<none>" && !strings.Contains(analysisRuns, "Error:") {
		cli.Printf("\nWould you like to view details of an analysis run? [y/N]: ")
		var response string
		_, _ = fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			return showAnalysisRunSelection(rollout)
		}
	}

	return nil
}

func showAnalysisRunSelection(rollout *unstructured.Unstructured) error {
	// Get the analysis runs for this rollout
	runs, err := getAnalysisRunsForSelection(rollout)
	if err != nil {
		return fmt.Errorf("failed to get analysis runs: %s", err)
	}

	if len(runs) == 0 {
		cli.Printf("No analysis runs found for this rollout\n")
		return nil
	}

	// Show selection menu
	selectedRun, err := analysisRunSelectMenu(runs)
	if err != nil {
		return fmt.Errorf("failed to select analysis run: %s", err)
	}

	if selectedRun == "" {
		cli.Printf("No analysis run selected\n")
		return nil
	}

	// Call the analysis get function
	return showAnalysisRunDetails(selectedRun, rollout.GetNamespace())
}

func getAnalysisRunsForSelection(rollout *unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	analysisRunGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysisruns",
	}

	rolloutName := rollout.GetName()
	rolloutNamespace := rollout.GetNamespace()

	// Get the current and stable pod hashes
	currentPodHash, _, _ := unstructured.NestedString(rollout.Object, "status", "currentPodHash")
	stablePodHash, _, _ := unstructured.NestedString(rollout.Object, "status", "stableRS")

	// Use a map to deduplicate runs by name
	runMap := make(map[string]unstructured.Unstructured)

	// Get unique pod hashes to avoid querying the same hash twice
	uniquePodHashes := make(map[string]bool)
	for _, podHash := range []string{currentPodHash, stablePodHash} {
		if podHash != "" && !uniquePodHashes[podHash] {
			uniquePodHashes[podHash] = true

			labelSelector := fmt.Sprintf("rollouts-pod-template-hash=%s", podHash)
			analysisRuns, err := dynamicClient.Resource(analysisRunGVR).Namespace(rolloutNamespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err == nil {
				for _, run := range analysisRuns.Items {
					runMap[run.GetName()] = run
				}
			}
		}
	}

	// Fallback to owner reference check if no runs found
	if len(runMap) == 0 {
		analysisRuns, err := dynamicClient.Resource(analysisRunGVR).Namespace(rolloutNamespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, run := range analysisRuns.Items {
				if isOwnedByRollout(&run, rolloutName) {
					runMap[run.GetName()] = run
				}
			}
		}
	}

	// Convert map back to slice
	var allRuns []unstructured.Unstructured
	for _, run := range runMap {
		allRuns = append(allRuns, run)
	}

	return allRuns, nil
}

func analysisRunSelectMenu(runs []unstructured.Unstructured) (string, error) {
	if len(runs) == 0 {
		return "", fmt.Errorf("no analysis runs available")
	}

	var runOptions []string
	for _, run := range runs {
		name := run.GetName()
		phase, _, _ := unstructured.NestedString(run.Object, "status", "phase")
		if phase == "" {
			phase = "Unknown"
		}

		// Get pod template hash to show revision
		runLabels := run.GetLabels()
		revisionInfo := ""
		if runLabels != nil {
			if podHash := runLabels["rollouts-pod-template-hash"]; podHash != "" {
				revisionInfo = fmt.Sprintf(" [%s]", podHash[:8]) // Show first 8 chars of hash
			}
		}

		// Get age
		creationTime := run.GetCreationTimestamp()
		age := "unknown"
		if !creationTime.IsZero() {
			age = helper.GetAge(creationTime.Time)
		}

		option := fmt.Sprintf("%s [%s]%s (age:%s)", name, phase, revisionInfo, age)
		runOptions = append(runOptions, option)
	}

	// Add "None" option to allow user to cancel
	runOptions = append(runOptions, "None - Cancel")

	var selectedOption string
	prompt := &survey.Select{
		Message: "Select an analysis run to view details:",
		Options: runOptions,
	}

	err := survey.AskOne(prompt, &selectedOption)
	if err != nil {
		return "", err
	}

	// Handle "None" selection
	if selectedOption == "None - Cancel" {
		return "", nil
	}

	// Extract run name from the selected option (everything before the first space)
	parts := strings.Split(selectedOption, " ")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("invalid selection")
}

func showAnalysisRunDetails(runName, namespace string) error {
	// This is equivalent to calling: yak rollouts analysis get -r <runName>
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to connect to Kubernetes cluster: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to create dynamic client: %s", err)
	}

	runGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysisruns",
	}

	run, err := dynamicClient.Resource(runGVR).Namespace(namespace).Get(context.Background(), runName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("analysis run %s not found in namespace %s: %s", runName, namespace, err)
	}

	cli.Printf("\n%s\n", strings.Repeat("=", 60))
	cli.Printf("ANALYSIS RUN DETAILS: %s\n", runName)
	cli.Printf("%s\n", strings.Repeat("=", 60))

	// Use the same formatting function from analysis.go
	formatAnalysisRunDetails(run)

	return nil
}

func formatStrategy(rollout *unstructured.Unstructured) string {
	parts := []string{}

	// Check for canary strategy
	if canarySpec, found, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "canary"); found && canarySpec != nil {
		parts = append(parts, "type:Canary")

		if maxSurge, found, _ := unstructured.NestedString(rollout.Object, "spec", "strategy", "canary", "maxSurge"); found {
			parts = append(parts, "maxSurge:"+maxSurge)
		}
		if maxUnavailable, found, _ := unstructured.NestedString(rollout.Object, "spec", "strategy", "canary", "maxUnavailable"); found {
			parts = append(parts, "maxUnavailable:"+maxUnavailable)
		}
		if steps, found, _ := unstructured.NestedSlice(rollout.Object, "spec", "strategy", "canary", "steps"); found {
			parts = append(parts, fmt.Sprintf("steps:%d", len(steps)))
		}
	} else if blueGreenSpec, found, _ := unstructured.NestedMap(rollout.Object, "spec", "strategy", "blueGreen"); found && blueGreenSpec != nil {
		parts = append(parts, "type:BlueGreen")

		if activeService, found, _ := unstructured.NestedString(rollout.Object, "spec", "strategy", "blueGreen", "activeService"); found {
			parts = append(parts, "activeService:"+activeService)
		}
		if previewService, found, _ := unstructured.NestedString(rollout.Object, "spec", "strategy", "blueGreen", "previewService"); found {
			parts = append(parts, "previewService:"+previewService)
		}
	}

	if len(parts) == 0 {
		return "<none>"
	}

	return strings.Join(parts, "\n")
}

func formatRevision(rollout *unstructured.Unstructured) string {
	parts := []string{}

	// Get current revision from rollout annotations (this is how argo rollouts tracks revisions)
	annotations := rollout.GetAnnotations()
	if annotations != nil {
		if revisionStr, exists := annotations["rollout.argoproj.io/revision"]; exists && revisionStr != "" {
			parts = append(parts, "current:"+revisionStr)
		}
	}

	// Get stable revision from ReplicaSets
	stableRevision := getStableRevision(rollout)
	if stableRevision != "" {
		parts = append(parts, "stable:"+stableRevision)
	}

	// Fallback to generation if no revision found
	if len(parts) == 0 {
		if generation, found, _ := unstructured.NestedInt64(rollout.Object, "metadata", "generation"); found {
			parts = append(parts, fmt.Sprintf("generation:%d", generation))
		}
	}

	if len(parts) == 0 {
		return "<none>"
	}

	return strings.Join(parts, "\n")
}

func getStableRevision(rollout *unstructured.Unstructured) string {
	// Get the stable RS hash from rollout status
	stableRS, found, _ := unstructured.NestedString(rollout.Object, "status", "stableRS")
	if !found || stableRS == "" {
		return ""
	}

	// Try to find the ReplicaSet with this hash and get its revision
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return ""
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return ""
	}

	rsGVR := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "replicasets",
	}

	// List ReplicaSets in the same namespace
	rsList, err := dynamicClient.Resource(rsGVR).Namespace(rollout.GetNamespace()).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return ""
	}

	// Find the ReplicaSet with the matching pod template hash
	for _, rs := range rsList.Items {
		rsLabels := rs.GetLabels()
		if rsLabels != nil {
			if podHash, exists := rsLabels["pod-template-hash"]; exists && podHash == stableRS {
				// Get revision from annotations
				annotations := rs.GetAnnotations()
				if annotations != nil {
					if revisionStr, exists := annotations["rollout.argoproj.io/revision"]; exists && revisionStr != "" {
						return revisionStr
					}
					// Fallback to deployment revision annotation
					if revisionStr, exists := annotations["deployment.kubernetes.io/revision"]; exists && revisionStr != "" {
						return revisionStr
					}
				}
			}
		}
	}

	return ""
}

func formatConditions(rollout *unstructured.Unstructured) string {
	conditions, found, _ := unstructured.NestedSlice(rollout.Object, "status", "conditions")
	if !found || len(conditions) == 0 {
		return "<none>"
	}

	// Parse all conditions first
	conditionMap := make(map[string]string)
	for _, condition := range conditions {
		condMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _ := condMap["type"].(string)
		status, _ := condMap["status"].(string)

		if condType != "" {
			conditionMap[condType] = status
		}
	}

	// Filter and prioritize conditions to show only relevant ones
	var relevantConditions []string

	// Priority order for display (most important first)
	priorityConditions := []string{
		"Healthy",
		"Completed",
		"Degraded",
		"Paused",
		"Progressing",
		"Available",
		"ReplicaFailure",
	}

	// Apply logic to show only relevant conditions
	for _, condType := range priorityConditions {
		if status, exists := conditionMap[condType]; exists {
			// Skip Progressing:True if Completed:True or Healthy:True
			if condType == "Progressing" && status == "True" {
				if (conditionMap["Completed"] == "True") || (conditionMap["Healthy"] == "True") {
					continue
				}
			}

			// Skip Available:True if Healthy:True (redundant)
			if condType == "Available" && status == "True" {
				if conditionMap["Healthy"] == "True" {
					continue
				}
			}

			// Only show False conditions for certain types that matter when False
			if status == "False" && (condType == "Healthy" || condType == "Available" || condType == "Completed") {
				continue // Don't show Healthy:False, Available:False, Completed:False as they're not useful
			}

			// Only show ReplicaFailure if it's True (failure condition)
			if condType == "ReplicaFailure" && status != "True" {
				continue
			}

			condStr := fmt.Sprintf("%s:%s", condType, status)
			relevantConditions = append(relevantConditions, condStr)
		}
	}

	if len(relevantConditions) == 0 {
		return "<none>"
	}

	return strings.Join(relevantConditions, "\n")
}

func formatAnalysisRuns(rollout *unstructured.Unstructured) string {
	// First check if there are any current analysis runs in the rollout status
	currentRuns := getCurrentAnalysisRuns(rollout)
	if len(currentRuns) > 0 {
		return strings.Join(currentRuns, "\n")
	}

	// If no current runs, query for associated AnalysisRun resources
	associatedRuns, err := getAssociatedAnalysisRuns(rollout)
	if err != nil {
		return fmt.Sprintf("Error: %s", err.Error())
	}

	if len(associatedRuns) == 0 {
		return "<none>"
	}

	return strings.Join(associatedRuns, "\n")
}

func getCurrentAnalysisRuns(rollout *unstructured.Unstructured) []string {
	parts := []string{}

	// Check multiple locations for current analysis runs
	locations := [][]string{
		{"status", "canary", "currentBackgroundAnalysisRuns"},
		{"status", "canary", "currentStepAnalysisRuns"},
		{"status", "blueGreen", "prePromotionAnalysisRuns"},
		{"status", "blueGreen", "postPromotionAnalysisRuns"},
		{"status", "analysisRuns"},
	}

	for _, location := range locations {
		analysisRuns, found, _ := unstructured.NestedSlice(rollout.Object, location...)
		if found && len(analysisRuns) > 0 {
			for _, run := range analysisRuns {
				runMap, ok := run.(map[string]interface{})
				if !ok {
					continue
				}

				name, _ := runMap["name"].(string)
				status, _ := runMap["status"].(string)
				phase, _ := runMap["phase"].(string)

				if name != "" {
					runStr := name
					if status != "" {
						runStr += fmt.Sprintf(":%s", status)
					}
					if phase != "" {
						runStr += fmt.Sprintf(" (%s)", phase)
					}
					parts = append(parts, runStr)
				}
			}
		}
	}

	return parts
}

func getAssociatedAnalysisRuns(rollout *unstructured.Unstructured) ([]string, error) {
	config, _, err := helper.InitKubeClusterConfig()
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	analysisRunGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "analysisruns",
	}

	rolloutName := rollout.GetName()
	rolloutNamespace := rollout.GetNamespace()

	// Get the current pod template hash from the rollout status
	currentPodHash, _, _ := unstructured.NestedString(rollout.Object, "status", "currentPodHash")
	stablePodHash, _, _ := unstructured.NestedString(rollout.Object, "status", "stableRS")

	// Try multiple strategies to find relevant analysis runs
	var allRuns []unstructured.Unstructured

	// Strategy 1: Use rollouts-pod-template-hash label to match current revision
	if currentPodHash != "" {
		labelSelector := fmt.Sprintf("rollouts-pod-template-hash=%s", currentPodHash)
		analysisRuns, err := dynamicClient.Resource(analysisRunGVR).Namespace(rolloutNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			allRuns = append(allRuns, analysisRuns.Items...)
		}
	}

	// Strategy 2: Use rollouts-pod-template-hash label to match stable revision
	if stablePodHash != "" && stablePodHash != currentPodHash {
		labelSelector := fmt.Sprintf("rollouts-pod-template-hash=%s", stablePodHash)
		analysisRuns, err := dynamicClient.Resource(analysisRunGVR).Namespace(rolloutNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			allRuns = append(allRuns, analysisRuns.Items...)
		}
	}

	// Strategy 3: Fallback to owner reference check if no pod hashes found
	if len(allRuns) == 0 {
		analysisRuns, err := dynamicClient.Resource(analysisRunGVR).Namespace(rolloutNamespace).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			for _, run := range analysisRuns.Items {
				if isOwnedByRollout(&run, rolloutName) {
					allRuns = append(allRuns, run)
				}
			}
		}
	}

	// Remove duplicates and format results
	seen := make(map[string]bool)
	var results []string

	for _, run := range allRuns {
		name := run.GetName()
		if seen[name] {
			continue
		}
		seen[name] = true

		phase, _, _ := unstructured.NestedString(run.Object, "status", "phase")
		if phase == "" {
			phase = "Unknown"
		}

		// Get the pod template hash from the analysis run labels to show which revision it belongs to
		runLabels := run.GetLabels()
		podHash := ""
		if runLabels != nil {
			podHash = runLabels["rollouts-pod-template-hash"]
		}

		// Get creation time
		creationTime := run.GetCreationTimestamp()
		age := "unknown"
		if !creationTime.IsZero() {
			age = helper.GetAge(creationTime.Time)
		}

		runStr := fmt.Sprintf("%s:%s", name, phase)
		if podHash != "" {
			// Indicate which revision this analysis run belongs to
			var revisionType string
			switch podHash {
			case currentPodHash:
				revisionType = "current"
			case stablePodHash:
				revisionType = "stable"
			default:
				revisionType = "unknown"
			}
			runStr += fmt.Sprintf(" (%s rev)", revisionType)
		}
		runStr += fmt.Sprintf(" (age:%s)", age)

		results = append(results, runStr)
	}

	return results, nil
}

func isOwnedByRollout(run *unstructured.Unstructured, rolloutName string) bool {
	ownerRefs, found, _ := unstructured.NestedSlice(run.Object, "metadata", "ownerReferences")
	if !found {
		return false
	}

	for _, ref := range ownerRefs {
		refMap, ok := ref.(map[string]interface{})
		if !ok {
			continue
		}

		kind, _ := refMap["kind"].(string)
		name, _ := refMap["name"].(string)

		if kind == "Rollout" && name == rolloutName {
			return true
		}
	}

	return false
}

var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get detailed information about a rollout",
	Example: `yak rollouts get -r my-rollout`,
	RunE:    get,
}

func init() {
	getCmd.Flags().StringVarP(&providedGetFlags.rollout, "rollout", "r", "", "Rollout name (required)")
	_ = getCmd.MarkFlagRequired("rollout")
}
