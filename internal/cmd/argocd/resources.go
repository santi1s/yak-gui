package argocd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/doctolib/yak/cli"
	argocdhelper "github.com/doctolib/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

// truncateString truncates a string to a specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type ArgoCDResourcesFlags struct {
	application string
	kind        string
	namespace   string
	paginate    bool
	pageSize    int
}

var providedResourcesFlags ArgoCDResourcesFlags

type ResourceInfo struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Group     string `json:"group"`
	Version   string `json:"version"`
	Health    string `json:"health"`
	Status    string `json:"status"`
}

func resources(cmd *cobra.Command, args []string) error {
	appName := providedResourcesFlags.application

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

	// Get the resource tree for the application
	resourceTree, err := apiclient.AppClient.ResourceTree(context.Background(), &application.ResourcesQuery{
		ApplicationName: &myApp.Name,
		AppNamespace:    &myApp.Namespace,
	})
	if err != nil {
		return fmt.Errorf("failed to get resource tree for application %s: %s", appName, err)
	}

	// Convert to our resource info structure
	var allResources []ResourceInfo

	// Add managed resources
	for _, node := range resourceTree.Nodes {
		health := "Unknown"
		if node.Health != nil {
			health = string(node.Health.Status)
			if health == "" {
				health = "Unknown"
			}
		}

		resource := ResourceInfo{
			Name:      node.Name,
			Kind:      node.Kind,
			Namespace: node.Namespace,
			Group:     node.Group,
			Version:   node.Version,
			Health:    health,
			Status:    "Managed",
		}

		// Apply filters
		if shouldIncludeResource(resource) {
			allResources = append(allResources, resource)
		}
	}

	// Add orphaned resources
	for _, orphan := range resourceTree.OrphanedNodes {
		resource := ResourceInfo{
			Name:      orphan.Name,
			Kind:      orphan.Kind,
			Namespace: orphan.Namespace,
			Group:     orphan.Group,
			Version:   orphan.Version,
			Health:    "Unknown",
			Status:    "Orphaned",
		}

		// Apply filters
		if shouldIncludeResource(resource) {
			allResources = append(allResources, resource)
		}
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(allResources)
	}

	// Format output for console
	formatResourcesOutput(allResources, appName)
	return nil
}

func shouldIncludeResource(resource ResourceInfo) bool {
	// Filter by kind if specified
	if providedResourcesFlags.kind != "" &&
		!strings.EqualFold(resource.Kind, providedResourcesFlags.kind) {
		return false
	}

	// Filter by namespace if specified
	if providedResourcesFlags.namespace != "" &&
		!strings.EqualFold(resource.Namespace, providedResourcesFlags.namespace) {
		return false
	}

	return true
}

// ResourcesDisplayer implements the PaginationDisplayer interface for resources output
type ResourcesDisplayer struct {
	resources []ResourceInfo
	appName   string
}

func (r *ResourcesDisplayer) DisplayPage(start, end int) {
	cli.Printf("%-25s %-20s %-15s %-15s %-12s %s\n", "Name", "Kind", "Namespace", "Health", "Status", "Group/Version")
	cli.Println("--------------------------------------------------------------------------------------------------")

	for i := start; i < end && i < len(r.resources); i++ {
		resource := r.resources[i]
		namespace := resource.Namespace
		if namespace == "" {
			namespace = "<cluster>"
		}

		groupVersion := ""
		if resource.Group != "" {
			groupVersion = resource.Group + "/" + resource.Version
		} else {
			groupVersion = resource.Version
		}

		cli.Printf("%-25s %-20s %-15s %-15s %-12s %s\n",
			truncateString(resource.Name, 25),
			resource.Kind,
			namespace,
			resource.Health,
			resource.Status,
			groupVersion)
	}
}

func (r *ResourcesDisplayer) GetTotal() int {
	return len(r.resources)
}

func (r *ResourcesDisplayer) DisplaySummary() {
	managedCount := 0
	orphanedCount := 0
	for _, resource := range r.resources {
		if resource.Status == "Managed" {
			managedCount++
		} else {
			orphanedCount++
		}
	}

	cli.Printf("\nSummary: %d managed, %d orphaned resources\n", managedCount, orphanedCount)
}

func formatResourcesOutput(resources []ResourceInfo, appName string) {
	if len(resources) == 0 {
		cli.Printf("No resources found for application %s\n", appName)
		return
	}

	// Sort resources by kind, then by name
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].Kind != resources[j].Kind {
			return resources[i].Kind < resources[j].Kind
		}
		return resources[i].Name < resources[j].Name
	})

	cli.Printf("** Resources for Application: %s **\n", appName)

	displayer := &ResourcesDisplayer{
		resources: resources,
		appName:   appName,
	}

	if providedResourcesFlags.paginate {
		argocdhelper.PaginateOutput(displayer, providedResourcesFlags.pageSize)
	} else {
		displayer.DisplayPage(0, len(resources))
		displayer.DisplaySummary()
	}
}

var resourcesCmd = &cobra.Command{
	Use:     "resources",
	Short:   "List all resources managed by an ArgoCD application",
	Example: "yak argocd resources --application my-app\nyak argocd resources --application my-app --kind Deployment",
	RunE:    resources,
}

func init() {
	resourcesCmd.Flags().StringVarP(&providedResourcesFlags.application, "application", "a", "", "ArgoCD application name")
	resourcesCmd.Flags().StringVarP(&providedResourcesFlags.kind, "kind", "k", "", "Filter by resource kind")
	resourcesCmd.Flags().StringVarP(&providedResourcesFlags.namespace, "namespace", "n", "", "Filter by namespace")
	resourcesCmd.Flags().BoolVar(&providedResourcesFlags.paginate, "paginate", false, "Enable pagination for large resource lists")
	resourcesCmd.Flags().IntVar(&providedResourcesFlags.pageSize, "page-size", 20, "Number of resources per page (default: 20)")
	_ = resourcesCmd.MarkFlagRequired("application")
}
