package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/util/argo/normalizers"
	gitopsdiff "github.com/argoproj/gitops-engine/pkg/diff"
	"github.com/santi1s/yak/cli"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type ArgoCDDiffFlags struct {
	application   string
	resourcesOnly bool
	resource      string
}

var providedDiffFlags ArgoCDDiffFlags

func diff(cmd *cobra.Command, args []string) error {
	appName := providedDiffFlags.application

	// Explicitly ignore global JSON/YAML flags for diff command
	if providedFlags.json || providedFlags.yaml {
		return fmt.Errorf("JSON/YAML output is not supported for diff command - use interactive mode instead")
	}

	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	myApp, err := argocdhelper.GetApplication(apiclient.AppClient, appName, providedFlags.project)
	if err != nil {
		return err
	}

	// Get the managed resources diff from ArgoCD
	diffResult, err := apiclient.AppClient.ManagedResources(context.Background(), &application.ResourcesQuery{
		ApplicationName: &myApp.Name,
		AppNamespace:    &myApp.Namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to get managed resources for application %s: %s", appName, err)
	}

	// Handle different output modes
	if providedDiffFlags.resourcesOnly {
		return resourcesOnlyOutput(diffResult, appName, myApp.Status.Sync.Status, myApp)
	} else if providedDiffFlags.resource != "" {
		return singleResourceDiffOutput(diffResult, appName, myApp.Status.Sync.Status, myApp, providedDiffFlags.resource)
	} else {
		// Format output for console with pagination (default behavior)
		return paginatedDiffOutput(diffResult, appName, myApp.Status.Sync.Status, myApp)
	}
}

func formatActualDiffOutput(managedResources *application.ManagedResourcesResponse, appName string, appSyncStatus v1alpha1.SyncStatusCode, myApp *v1alpha1.Application) {
	if managedResources == nil || len(managedResources.Items) == 0 {
		cli.Printf("No managed resources found for application %s\n", appName)
		return
	}

	cli.Printf("===== Differences for Application: %s (Status: %s) =====\n", appName, appSyncStatus)

	// If the application is synced according to ArgoCD, be much more conservative
	// about what we consider "differences"
	isSynced := appSyncStatus == v1alpha1.SyncStatusCodeSynced

	hasChanges := false
	outOfSyncResources := 0

	for _, item := range managedResources.Items {
		// Check if this resource is actually out of sync according to ArgoCD
		// The ResourceDiff API includes sync status information
		isOutOfSync := false

		// A resource is considered out of sync if:
		// 1. It has differences between target and normalized live state AND
		// 2. ArgoCD hasn't marked it as synced (which would be in the sync status)
		// Use ArgoCD's official diff logic instead of custom implementation
		if item.NormalizedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.NormalizedLiveState, myApp)
		} else if item.PredictedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.PredictedLiveState, myApp)
		} else if item.LiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.LiveState, myApp)
		}

		if isOutOfSync {
			outOfSyncResources++
			// Only show diff for resources that have actual meaningful differences
			// If the app is synced, be even more conservative
			if shouldShowResourceDiff(item, isSynced, myApp) {
				hasChanges = true
				showResourceDiff(item)
			}
		}
	}

	if !hasChanges {
		if outOfSyncResources == 0 {
			cli.Printf("\n✅ No differences found. Application is in sync.\n")
		} else {
			cli.Printf("\n✅ Found %d resources with technical differences, but they appear to be \n", outOfSyncResources)
			cli.Printf("    metadata-only changes that ArgoCD ignores for sync determination.\n")
			cli.Printf("    This matches ArgoCD's assessment that the application is '%s'.\n", appSyncStatus)
		}
	} else {
		if isSynced {
			// If ArgoCD says it's synced, trust it and don't show differences
			cli.Printf("\n✅ Application is reported as 'Synced' by ArgoCD.\n")
			cli.Printf("    Found %d resources with differences, but ArgoCD considers them non-significant.\n", outOfSyncResources)
			cli.Printf("    These are likely metadata-only changes or fields that ArgoCD ignores.\n")
		} else {
			cli.Printf("\n🔄 Found resources with meaningful differences.\n")
		}
	}
}

func shouldShowResourceDiff(item *v1alpha1.ResourceDiff, appIsSynced bool, app *v1alpha1.Application) bool {
	// If ArgoCD says the app is synced, be extremely conservative
	// Trust ArgoCD's sync determination over our diff analysis
	if appIsSynced {
		return false
	}

	// For non-synced apps, use ArgoCD's official diff logic
	if item.NormalizedLiveState != "" && item.TargetState != "" {
		return isResourceOutOfSync(item.TargetState, item.NormalizedLiveState, app)
	} else if item.PredictedLiveState != "" && item.TargetState != "" {
		return isResourceOutOfSync(item.TargetState, item.PredictedLiveState, app)
	} else if item.LiveState != "" && item.TargetState != "" {
		return isResourceOutOfSync(item.TargetState, item.LiveState, app)
	}

	return true
}

func showResourceDiff(item *v1alpha1.ResourceDiff) {
	// Extract resource info
	resourceName := "unknown"
	resourceKind := "unknown"
	resourceNamespace := ""

	if item.Group != "" {
		resourceKind = fmt.Sprintf("%s %s", item.Group, item.Kind)
	} else {
		resourceKind = item.Kind
	}

	if item.Name != "" {
		resourceName = item.Name
	}

	if item.Namespace != "" {
		resourceNamespace = item.Namespace
	}

	// Print resource header
	cli.Printf("\n--- %s: %s", resourceKind, resourceName)
	if resourceNamespace != "" {
		cli.Printf(" (namespace: %s)", resourceNamespace)
	}
	cli.Printf(" ---\n")

	// Show the diff - prioritize normalized comparison like ArgoCD does
	// Clean the states before showing diff
	var cleanTarget, cleanLive string
	resourceFullName := fmt.Sprintf("%s/%s", resourceKind, resourceName)
	if item.NormalizedLiveState != "" && item.TargetState != "" {
		cli.Printf("📄 Normalized state differences (ArgoCD sync comparison):\n")
		cleanTarget = cleanResourceState(item.TargetState)
		cleanLive = cleanResourceState(item.NormalizedLiveState)
		showUnifiedDiff(cleanTarget, cleanLive, resourceFullName)
	} else if item.PredictedLiveState != "" && item.TargetState != "" {
		cli.Printf("📄 Predicted changes:\n")
		cleanTarget = cleanResourceState(item.TargetState)
		cleanLive = cleanResourceState(item.PredictedLiveState)
		showUnifiedDiff(cleanTarget, cleanLive, resourceFullName)
	} else if item.LiveState != "" && item.TargetState != "" {
		cli.Printf("📄 Raw state differences (includes k8s metadata):\n")
		cleanTarget = cleanResourceState(item.TargetState)
		cleanLive = cleanResourceState(item.LiveState)
		showUnifiedDiff(cleanTarget, cleanLive, resourceFullName)
	}
}

// generateUnifiedDiff creates a git-style unified diff between two strings
func generateUnifiedDiff(target, live, resourceName string) []string {
	// Convert JSON to pretty-printed YAML for better readability
	targetYAML := jsonToYAML(target)
	liveYAML := jsonToYAML(live)

	targetLines := strings.Split(targetYAML, "\n")
	liveLines := strings.Split(liveYAML, "\n")

	// Switch the order: live first (what we have), target second (what we want)
	// This makes target (source of truth) show as green (+) and live (current) show as red (-)
	return diffLines(liveLines, targetLines, resourceName)
}

// jsonToYAML converts JSON to pretty YAML
func jsonToYAML(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}

	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return jsonStr // Return original if parsing fails
	}

	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return jsonStr // Return original if marshaling fails
	}

	return string(yamlBytes)
}

// diffLines creates a simple unified diff between two sets of lines
func diffLines(targetLines, liveLines []string, resourceName string) []string {
	var result []string

	// Add header - now that we switched the order:
	// --- represents what we currently have (live state)
	// +++ represents what we want (target state from Git)
	result = append(result, fmt.Sprintf("--- a/%s (current cluster state)", resourceName))
	result = append(result, fmt.Sprintf("+++ b/%s (desired git state)", resourceName))

	// Simple line-by-line comparison
	maxLines := len(targetLines)
	if len(liveLines) > maxLines {
		maxLines = len(liveLines)
	}

	contextLines := 3
	inDiff := false
	hunkStart := -1

	for i := 0; i < maxLines; i++ {
		targetLine := ""
		liveLine := ""

		if i < len(targetLines) {
			targetLine = targetLines[i]
		}
		if i < len(liveLines) {
			liveLine = liveLines[i]
		}

		if targetLine != liveLine {
			if !inDiff {
				// Start of a new hunk
				hunkStart = maxInt(0, i-contextLines)
				inDiff = true

				// Add hunk header
				result = append(result, fmt.Sprintf("@@ -%d,%d +%d,%d @@",
					hunkStart+1, minInt(len(targetLines), i+contextLines*2+1)-hunkStart,
					hunkStart+1, minInt(len(liveLines), i+contextLines*2+1)-hunkStart))

				// Add context lines before
				for j := hunkStart; j < i; j++ {
					if j < len(targetLines) {
						result = append(result, " "+targetLines[j])
					}
				}
			}

			// Add the differing lines
			if i < len(targetLines) {
				result = append(result, "-"+targetLine)
			}
			if i < len(liveLines) {
				result = append(result, "+"+liveLine)
			}
		} else if inDiff {
			// Add context line
			result = append(result, " "+targetLine)

			// Check if we should end the hunk
			if i >= len(targetLines)-1 && i >= len(liveLines)-1 {
				inDiff = false
			}
		}
	}

	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func showUnifiedDiff(target, live, resourceName string) {
	diffLines := generateUnifiedDiff(target, live, resourceName)

	if len(diffLines) <= 2 { // Only headers means no differences
		cli.Printf("✅ No differences found\n")
		return
	}

	for _, line := range diffLines {
		// Color code the diff lines
		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			cli.Printf("\033[1m%s\033[0m\n", line) // Bold
		case strings.HasPrefix(line, "@@"):
			cli.Printf("\033[36m%s\033[0m\n", line) // Cyan
		case strings.HasPrefix(line, "-"):
			cli.Printf("\033[31m%s\033[0m\n", line) // Red
		case strings.HasPrefix(line, "+"):
			cli.Printf("\033[32m%s\033[0m\n", line) // Green
		default:
			cli.Printf("%s\n", line) // Normal
		}
	}
}

// DiffPaginator implements the PaginationDisplayer interface for diff output
type DiffPaginator struct {
	items       []*v1alpha1.ResourceDiff
	appName     string
	syncStatus  v1alpha1.SyncStatusCode
	app         *v1alpha1.Application
	diffResults [][]string // Pre-computed diff results for each resource
}

func (dp *DiffPaginator) DisplayPage(start, end int) {
	for i := start; i < end && i < len(dp.items); i++ {
		item := dp.items[i]

		// Extract resource info
		resourceName := "unknown"
		resourceKind := "unknown"
		resourceNamespace := ""

		if item.Group != "" {
			resourceKind = fmt.Sprintf("%s %s", item.Group, item.Kind)
		} else {
			resourceKind = item.Kind
		}

		if item.Name != "" {
			resourceName = item.Name
		}

		if item.Namespace != "" {
			resourceNamespace = item.Namespace
		}

		// Print resource header
		cli.Printf("\n--- %s: %s", resourceKind, resourceName)
		if resourceNamespace != "" {
			cli.Printf(" (namespace: %s)", resourceNamespace)
		}
		cli.Printf(" ---\n")

		// Show the pre-computed diff
		if i < len(dp.diffResults) && len(dp.diffResults[i]) > 0 {
			for _, line := range dp.diffResults[i] {
				// Color code the diff lines
				switch {
				case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
					cli.Printf("\033[1m%s\033[0m\n", line) // Bold
				case strings.HasPrefix(line, "@@"):
					cli.Printf("\033[36m%s\033[0m\n", line) // Cyan
				case strings.HasPrefix(line, "-"):
					cli.Printf("\033[31m%s\033[0m\n", line) // Red
				case strings.HasPrefix(line, "+"):
					cli.Printf("\033[32m%s\033[0m\n", line) // Green
				default:
					cli.Printf("%s\n", line) // Normal
				}
			}
		} else {
			cli.Printf("✅ No differences found\n")
		}
	}
}

func (dp *DiffPaginator) GetTotal() int {
	return len(dp.items)
}

func (dp *DiffPaginator) DisplaySummary() {
	outOfSyncResources := 0
	for i := range dp.diffResults {
		if len(dp.diffResults[i]) > 2 { // More than just headers
			outOfSyncResources++
		}
	}

	if outOfSyncResources == 0 {
		cli.Printf("\n✅ Summary: No differences found. Application is in sync.\n")
	} else {
		cli.Printf("\n🔄 Summary: Found %d resources with meaningful differences.\n", outOfSyncResources)
	}
}

// paginatedDiffOutput creates a paginated view of the diff results
func paginatedDiffOutput(managedResources *application.ManagedResourcesResponse, appName string, appSyncStatus v1alpha1.SyncStatusCode, myApp *v1alpha1.Application) error {
	if managedResources == nil || len(managedResources.Items) == 0 {
		cli.Printf("No managed resources found for application %s\n", appName)
		return nil
	}

	cli.Printf("===== Differences for Application: %s (Status: %s) =====\n", appName, appSyncStatus)

	// Pre-compute all diff results
	var diffItems []*v1alpha1.ResourceDiff
	var diffResults [][]string

	isSynced := appSyncStatus == v1alpha1.SyncStatusCodeSynced

	for _, item := range managedResources.Items {
		// Check if this resource is actually out of sync
		isOutOfSync := false
		if item.NormalizedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.NormalizedLiveState, myApp)
		} else if item.PredictedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.PredictedLiveState, myApp)
		} else if item.LiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.LiveState, myApp)
		}

		if isOutOfSync && shouldShowResourceDiff(item, isSynced, myApp) {
			// Generate diff for this resource
			var cleanTarget, cleanLive string
			resourceFullName := fmt.Sprintf("%s/%s", item.Kind, item.Name)

			if item.NormalizedLiveState != "" && item.TargetState != "" {
				cleanTarget = cleanResourceState(item.TargetState)
				cleanLive = cleanResourceState(item.NormalizedLiveState)
			} else if item.PredictedLiveState != "" && item.TargetState != "" {
				cleanTarget = cleanResourceState(item.TargetState)
				cleanLive = cleanResourceState(item.PredictedLiveState)
			} else if item.LiveState != "" && item.TargetState != "" {
				cleanTarget = cleanResourceState(item.TargetState)
				cleanLive = cleanResourceState(item.LiveState)
			}

			diffLines := generateUnifiedDiff(cleanTarget, cleanLive, resourceFullName)

			diffItems = append(diffItems, item)
			diffResults = append(diffResults, diffLines)
		}
	}

	if len(diffItems) == 0 {
		if isSynced {
			cli.Printf("\n✅ Application is reported as 'Synced' by ArgoCD.\n")
		} else {
			cli.Printf("\n✅ No differences found. Application is in sync.\n")
		}
		return nil
	}

	// Create paginator
	paginator := &DiffPaginator{
		items:       diffItems,
		appName:     appName,
		syncStatus:  appSyncStatus,
		app:         myApp,
		diffResults: diffResults,
	}

	// Use pagination with default page size of 1 for careful diff review
	pageSize := 1

	argocdhelper.PaginateOutput(paginator, pageSize)
	return nil
}

// resourcesOnlyOutput shows only the names of resources that have differences
func resourcesOnlyOutput(managedResources *application.ManagedResourcesResponse, appName string, appSyncStatus v1alpha1.SyncStatusCode, myApp *v1alpha1.Application) error {
	if managedResources == nil || len(managedResources.Items) == 0 {
		cli.Printf("No managed resources found for application %s\n", appName)
		return nil
	}

	cli.Printf("===== Resources with differences for Application: %s (Status: %s) =====\n", appName, appSyncStatus)

	isSynced := appSyncStatus == v1alpha1.SyncStatusCodeSynced
	outOfSyncCount := 0

	for _, item := range managedResources.Items {
		// Check if this resource is actually out of sync
		isOutOfSync := false
		if item.NormalizedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.NormalizedLiveState, myApp)
		} else if item.PredictedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.PredictedLiveState, myApp)
		} else if item.LiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.LiveState, myApp)
		}

		if isOutOfSync && shouldShowResourceDiff(item, isSynced, myApp) {
			outOfSyncCount++

			// Format as Kind/Name with optional namespace info
			kindName := fmt.Sprintf("%s/%s", item.Kind, item.Name)
			if item.Namespace != "" {
				kindName = fmt.Sprintf("%s/%s (namespace: %s)", item.Kind, item.Name, item.Namespace)
			}

			cli.Printf("- %s\n", kindName)
		}
	}

	if outOfSyncCount == 0 {
		if isSynced {
			cli.Printf("\n✅ Application is reported as 'Synced' by ArgoCD.\n")
		} else {
			cli.Printf("\n✅ No resources with meaningful differences found.\n")
		}
	} else {
		cli.Printf("\nTotal: %d resources with differences\n", outOfSyncCount)
		cli.Printf("\nUse --resource <Kind/Name> to see specific diff for a resource\n")
	}

	return nil
}

// parseResourceIdentifier parses a resource identifier in Kind/Name format
func parseResourceIdentifier(resource string) (string, string) {
	parts := strings.SplitN(resource, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	// If no slash, treat the whole thing as a name (for backward compatibility)
	return "", parts[0]
}

// singleResourceDiffOutput shows the diff for a specific resource only
func singleResourceDiffOutput(managedResources *application.ManagedResourcesResponse, appName string, appSyncStatus v1alpha1.SyncStatusCode, myApp *v1alpha1.Application, targetResource string) error {
	if managedResources == nil || len(managedResources.Items) == 0 {
		cli.Printf("No managed resources found for application %s\n", appName)
		return nil
	}

	cli.Printf("===== Diff for resource '%s' in Application: %s (Status: %s) =====\n", targetResource, appName, appSyncStatus)

	// Parse the resource identifier
	targetKind, targetName := parseResourceIdentifier(targetResource)

	isSynced := appSyncStatus == v1alpha1.SyncStatusCodeSynced
	found := false

	for _, item := range managedResources.Items {
		// Check if this is the resource we're looking for
		// Match by both Kind and Name if Kind is specified, otherwise just Name
		isMatch := false
		if targetKind != "" {
			isMatch = (item.Kind == targetKind && item.Name == targetName)
		} else {
			isMatch = (item.Name == targetName)
		}

		if !isMatch {
			continue
		}

		found = true

		// Check if this resource is actually out of sync
		isOutOfSync := false
		if item.NormalizedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.NormalizedLiveState, myApp)
		} else if item.PredictedLiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.PredictedLiveState, myApp)
		} else if item.LiveState != "" && item.TargetState != "" {
			isOutOfSync = isResourceOutOfSync(item.TargetState, item.LiveState, myApp)
		}

		if isOutOfSync && shouldShowResourceDiff(item, isSynced, myApp) {
			showResourceDiff(item)
		} else {
			// Extract resource info for display
			resourceKind := item.Kind
			if item.Group != "" {
				resourceKind = fmt.Sprintf("%s %s", item.Group, item.Kind)
			}

			cli.Printf("\n--- %s: %s ---\n", resourceKind, item.Name)
			cli.Printf("✅ No meaningful differences found for this resource\n")
		}
		break
	}

	if !found {
		cli.Printf("❌ Resource '%s' not found in application %s\n", targetResource, appName)
		cli.Printf("\nAvailable resources:\n")
		for _, item := range managedResources.Items {
			kindName := fmt.Sprintf("%s/%s", item.Kind, item.Name)
			if item.Namespace != "" {
				kindName = fmt.Sprintf("%s/%s (namespace: %s)", item.Kind, item.Name, item.Namespace)
			}
			cli.Printf("- %s\n", kindName)
		}
	}

	return nil
}

var diffCmd = &cobra.Command{
	Use:     "diff",
	Short:   "Show differences between Git and cluster state (same as ArgoCD)",
	Long:    "Show differences between Git manifests and live cluster state using the same field normalization logic as ArgoCD. This shows only the differences that ArgoCD considers significant for sync status determination. Uses git-style unified diff format with pagination (1 resource per page).",
	Example: "yak argocd diff -a my-app\nyak argocd diff -a my-app -l\nyak argocd diff -a my-app -r Deployment/my-deployment\nyak argocd diff --application my-app --resources-only\nyak argocd diff --application my-app --resource Deployment/my-deployment",
	RunE:    diff,
}

// cleanResourceState removes Kubernetes metadata from a JSON string
func cleanResourceState(jsonState string) string {
	if jsonState == "" || jsonState == "null" {
		return jsonState
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal([]byte(jsonState), &obj.Object); err != nil {
		return jsonState // Return original if we can't parse
	}

	// Additional null check after parsing
	if obj.Object == nil {
		return jsonState
	}

	removeKubernetesMetadata(&obj)

	cleanedBytes, err := json.Marshal(obj.Object)
	if err != nil {
		return jsonState // Return original if we can't marshal
	}

	return string(cleanedBytes)
}

// cleanManagedResourcesResponse removes Kubernetes metadata from all resource states
func cleanManagedResourcesResponse(response *application.ManagedResourcesResponse) *application.ManagedResourcesResponse {
	if response == nil {
		return response
	}

	// Create a copy to avoid modifying the original
	cleaned := &application.ManagedResourcesResponse{
		Items: make([]*v1alpha1.ResourceDiff, len(response.Items)),
	}

	for i, item := range response.Items {
		cleanedItem := &v1alpha1.ResourceDiff{
			Group:               item.Group,
			Kind:                item.Kind,
			Namespace:           item.Namespace,
			Name:                item.Name,
			ResourceVersion:     item.ResourceVersion,
			TargetState:         cleanResourceState(item.TargetState),
			LiveState:           cleanResourceState(item.LiveState),
			NormalizedLiveState: cleanResourceState(item.NormalizedLiveState),
			PredictedLiveState:  cleanResourceState(item.PredictedLiveState),
		}
		cleaned.Items[i] = cleanedItem
	}

	return cleaned
}

// removeKubernetesMetadata removes standard Kubernetes metadata fields that shouldn't affect sync status
func removeKubernetesMetadata(obj *unstructured.Unstructured) {
	if obj == nil || obj.Object == nil {
		return
	}

	// Remove standard Kubernetes metadata fields that are managed by the cluster
	metadata, found, err := unstructured.NestedMap(obj.Object, "metadata")
	if err != nil || !found {
		return
	}

	// Fields to remove from metadata
	fieldsToRemove := []string{
		"creationTimestamp",
		"managedFields",
		"resourceVersion",
		"uid",
		"generation",
		"selfLink",
	}

	for _, field := range fieldsToRemove {
		delete(metadata, field)
	}

	// Also clean up problematic annotations that cause false positives
	if annotations, found, err := unstructured.NestedMap(metadata, "annotations"); err == nil && found {
		// Remove the last-applied-configuration as it often has field ordering differences
		delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
		// Also remove other Kubernetes-managed annotations
		delete(annotations, "deployment.kubernetes.io/revision")

		// Update the metadata with cleaned annotations
		if len(annotations) > 0 {
			if err := unstructured.SetNestedMap(metadata, annotations, "annotations"); err != nil {
				// If we can't set annotations, just delete the field
				delete(metadata, "annotations")
			}
		} else {
			delete(metadata, "annotations")
		}
	}

	// Update the object
	if err := unstructured.SetNestedMap(obj.Object, metadata, "metadata"); err != nil {
		// If we can't set metadata, the object remains unchanged
		return
	}

	// Also remove status field as it's typically cluster-managed
	delete(obj.Object, "status")
}

// isResourceOutOfSync uses ArgoCD's official diff logic to determine if a resource is out of sync
func isResourceOutOfSync(targetState, liveState string, app *v1alpha1.Application) bool {
	if targetState == "" || liveState == "" {
		return targetState != liveState
	}

	// Check for null values which can cause panics
	if targetState == "null" || liveState == "null" {
		return targetState != liveState
	}

	// Parse target and live states
	var targetObj, liveObj unstructured.Unstructured

	if err := json.Unmarshal([]byte(targetState), &targetObj.Object); err != nil {
		if err := yaml.Unmarshal([]byte(targetState), &targetObj.Object); err != nil {
			return true // Parse error = assume different
		}
	}

	if err := json.Unmarshal([]byte(liveState), &liveObj.Object); err != nil {
		if err := yaml.Unmarshal([]byte(liveState), &liveObj.Object); err != nil {
			return true // Parse error = assume different
		}
	}

	// Additional null check after parsing
	if targetObj.Object == nil || liveObj.Object == nil {
		return (targetObj.Object == nil) != (liveObj.Object == nil)
	}

	// Remove Kubernetes metadata fields before comparison
	removeKubernetesMetadata(&targetObj)
	removeKubernetesMetadata(&liveObj)

	// Create ArgoCD normalizers (same as ArgoCD uses internally)
	normalizers := createArgoNormalizers(app)

	// Use ArgoCD's diff logic
	diffResult, err := gitopsdiff.Diff(&targetObj, &liveObj,
		gitopsdiff.WithNormalizer(normalizers),
	)

	if err != nil {
		return true // Diff error = assume different
	}

	return diffResult.Modified
}

// multiNormalizer combines multiple normalizers
type multiNormalizer struct {
	normalizers []gitopsdiff.Normalizer
}

func (n *multiNormalizer) Normalize(un *unstructured.Unstructured) error {
	for _, normalizer := range n.normalizers {
		if err := normalizer.Normalize(un); err != nil {
			return err
		}
	}
	return nil
}

// createArgoNormalizers creates the same normalizers that ArgoCD uses internally
func createArgoNormalizers(app *v1alpha1.Application) gitopsdiff.Normalizer {
	var normalizerList []gitopsdiff.Normalizer

	// Add known types normalizer (handles Kubernetes built-in types)
	if knownTypesNormalizer, err := normalizers.NewKnownTypesNormalizer(map[string]v1alpha1.ResourceOverride{}); err == nil {
		normalizerList = append(normalizerList, knownTypesNormalizer)
	}

	// Add ignore normalizer (handles application-specific ignore rules)
	var ignoreDifferences []v1alpha1.ResourceIgnoreDifferences
	if app != nil && app.Spec.IgnoreDifferences != nil {
		ignoreDifferences = app.Spec.IgnoreDifferences
	}

	if ignoreNormalizer, err := normalizers.NewIgnoreNormalizer(
		ignoreDifferences,
		map[string]v1alpha1.ResourceOverride{}, // Empty overrides for now
		normalizers.IgnoreNormalizerOpts{},
	); err == nil {
		normalizerList = append(normalizerList, ignoreNormalizer)
	}

	// Return our custom multi-normalizer
	return &multiNormalizer{normalizers: normalizerList}
}

func init() {
	diffCmd.Flags().StringVarP(&providedDiffFlags.application, "application", "a", "", "ArgoCD application name")
	diffCmd.Flags().BoolVarP(&providedDiffFlags.resourcesOnly, "resources-only", "l", false, "Show only resource names that have differences")
	diffCmd.Flags().StringVarP(&providedDiffFlags.resource, "resource", "r", "", "Show diff for a specific resource in Kind/Name format (e.g., Deployment/my-app)")
	_ = diffCmd.MarkFlagRequired("application")
}
