package argocd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var providedGetFlags ArgoCDGetFlags

type ArgoCDGetFlags struct {
	application string
}

func get(cmd *cobra.Command, args []string) error {
	appName := providedGetFlags.application
	if appName == "" {
		return fmt.Errorf("application name is required")
	}

	// Authentication
	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	// Get the project
	myProject, err := argocdhelper.GetArgoCDProject(apiclient.ProjectClient, providedFlags.project)
	if err != nil {
		return fmt.Errorf("project getting failed: %s", err)
	}

	// Get the application using the same pattern as status command
	apps, err := apiclient.AppClient.List(context.Background(), &application.ApplicationQuery{
		Project: []string{myProject.Name},
		Name:    &appName,
	})
	if err != nil {
		return fmt.Errorf("application %s list failed: %s", appName, err)
	}

	if len(apps.Items) == 0 {
		return fmt.Errorf("application %s not found in project %s", appName, myProject.Name)
	}

	app := &apps.Items[0]

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(app)
	} else {
		formatAppDetails(app)
	}
	return nil
}

func formatAppDetails(app *v1alpha1.Application) {
	// Format source details
	source := formatSource(app.Spec.Source)

	// Format destination details
	destination := formatDestination(app.Spec.Destination)

	// Format sync policy
	syncPolicy := formatSyncPolicy(app.Spec.SyncPolicy)

	// Format ignored differences
	ignoredDiffs := formatIgnoredDifferences(app.Spec.IgnoreDifferences)

	// Create table
	table := tablewriter.NewWriter(os.Stdout)

	// Configure table
	table.SetHeader([]string{"SOURCE", "DESTINATION", "SYNC POLICY", "IGNORED DIFFERENCES"})
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator(" ")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding(" ")
	table.SetNoWhiteSpace(true)
	table.SetAutoWrapText(true) // Enable text wrapping for long content
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetColMinWidth(0, 25) // SOURCE column
	table.SetColMinWidth(1, 20) // DESTINATION column
	table.SetColMinWidth(2, 30) // SYNC POLICY column (wider for full options)
	table.SetColMinWidth(3, 60) // IGNORED DIFFERENCES column (much wider for full JQ expressions)

	// Add data row
	table.Append([]string{source, destination, syncPolicy, ignoredDiffs})

	// Render table
	table.Render()
}

func formatSource(source *v1alpha1.ApplicationSource) string {
	if source == nil {
		return "<none>"
	}
	parts := []string{}

	if source.RepoURL != "" {
		// Extract repo name from URL
		repoName := source.RepoURL
		if strings.Contains(repoName, "/") {
			parts := strings.Split(repoName, "/")
			repoName = parts[len(parts)-1]
			repoName = strings.TrimSuffix(repoName, ".git")
		}
		parts = append(parts, "repo:"+repoName)
	}

	if source.Path != "" {
		parts = append(parts, "path:"+source.Path)
	}

	if source.TargetRevision != "" {
		parts = append(parts, "rev:"+source.TargetRevision)
	}

	if source.Chart != "" {
		parts = append(parts, "chart:"+source.Chart)
	}

	if len(parts) == 0 {
		return "<none>"
	}

	return strings.Join(parts, "\n")
}

func formatDestination(dest v1alpha1.ApplicationDestination) string {
	parts := []string{}

	if dest.Server != "" {
		server := dest.Server
		if server == "https://kubernetes.default.svc" {
			server = "in-cluster"
		}
		parts = append(parts, "server:"+server)
	}

	if dest.Name != "" {
		parts = append(parts, "cluster:"+dest.Name)
	}

	if dest.Namespace != "" {
		parts = append(parts, "namespace:"+dest.Namespace)
	}

	if len(parts) == 0 {
		return "<none>"
	}

	return strings.Join(parts, "\n")
}

func formatSyncPolicy(policy *v1alpha1.SyncPolicy) string {
	if policy == nil {
		return "<none>"
	}

	parts := []string{}

	if policy.Automated != nil {
		auto := []string{}
		if policy.Automated.Prune {
			auto = append(auto, "prune")
		}
		if policy.Automated.SelfHeal {
			auto = append(auto, "self-heal")
		}
		if len(auto) > 0 {
			parts = append(parts, "auto:"+strings.Join(auto, ","))
		} else {
			parts = append(parts, "auto:enabled")
		}
	} else {
		parts = append(parts, "manual")
	}

	if len(policy.SyncOptions) > 0 {
		options := strings.Join(policy.SyncOptions, ",")
		parts = append(parts, "opts:"+options)
	}

	return strings.Join(parts, "\n")
}

func formatIgnoredDifferences(diffs []v1alpha1.ResourceIgnoreDifferences) string {
	if len(diffs) == 0 {
		return "<none>"
	}

	parts := []string{}
	for i, diff := range diffs {
		// Resource identifier
		resourceStr := ""
		if diff.Group != "" {
			resourceStr += diff.Group + "/"
		}
		resourceStr += diff.Kind
		if diff.Name != "" {
			resourceStr += ":" + diff.Name
		}

		// Start with resource info
		var diffParts []string
		diffParts = append(diffParts, resourceStr)

		// Add JSON pointers (specific fields being ignored)
		if len(diff.JSONPointers) > 0 {
			diffParts = append(diffParts, "  Fields: "+strings.Join(diff.JSONPointers, ", "))
		}

		// Add JQ path expressions (if any) - show full content
		if len(diff.JQPathExpressions) > 0 {
			jqStr := strings.Join(diff.JQPathExpressions, "\n     ")
			diffParts = append(diffParts, "  JQ: "+jqStr)
		}

		// Add managed fields managers (if any)
		if len(diff.ManagedFieldsManagers) > 0 {
			diffParts = append(diffParts, "  Managers: "+strings.Join(diff.ManagedFieldsManagers, ", "))
		}

		// Join parts for this diff entry
		diffEntry := strings.Join(diffParts, "\n")

		// Add separator between multiple entries
		if i > 0 {
			diffEntry = "---\n" + diffEntry
		}

		parts = append(parts, diffEntry)
	}

	return strings.Join(parts, "\n")
}

var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get detailed information about an ArgoCD application",
	Example: `yak argocd get -a my-app`,
	RunE:    get,
}

func init() {
	getCmd.Flags().StringVarP(&providedGetFlags.application, "application", "a", "", "ArgoCD application name (required)")
	_ = getCmd.MarkFlagRequired("application")
}
