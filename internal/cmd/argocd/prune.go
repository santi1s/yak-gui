package argocd

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/santi1s/yak/cli"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

type ArgoCDPruneFlags struct {
	application string
	dryRun      bool
	confirm     bool
	verbose     bool
}

var providedPruneFlags ArgoCDPruneFlags

func prune(cmd *cobra.Command, args []string) error {
	appName := providedPruneFlags.application

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

	// Get orphaned resources for the application
	orphanedResources, err := argocdhelper.OrphanedResourcesArgoCD(apiclient.AppClient, providedFlags.project)
	if err != nil {
		return fmt.Errorf("failed to get orphaned resources for project %s: %s", providedFlags.project, err)
	}

	appOrphanedResources, exists := orphanedResources[appName]
	if !exists || len(appOrphanedResources) == 0 {
		cli.Printf("✅ No orphaned resources found for application %s\n", appName)
		return nil
	}

	if providedFlags.json || providedFlags.yaml {
		return formatOutput(appOrphanedResources)
	}

	// Display orphaned resources
	cli.Printf("===== Orphaned Resources for Application: %s =====\n", appName)
	cli.Printf("Found %d orphaned resources:\n\n", len(appOrphanedResources))

	for i, resource := range appOrphanedResources {
		cli.Printf("%d. %s %s", i+1, resource.Kind, resource.Name)
		if resource.Namespace != "" {
			cli.Printf(" (namespace: %s)", resource.Namespace)
		}
		if resource.Group != "" {
			cli.Printf(" [%s]", resource.Group)
		}

		// Show additional details in verbose mode
		if providedPruneFlags.verbose {
			cli.Printf("\n   📋 Full resource identifier: %s", formatResourceIdentifier(resource))
		}
		cli.Printf("\n")
	}

	if providedPruneFlags.dryRun {
		cli.Printf("\n🔍 Dry run: These resources would be deleted if --dry-run was not specified.\n")
		return nil
	}

	if !providedPruneFlags.confirm {
		cli.Printf("\n⚠️  To actually delete these orphaned resources, use --confirm flag.\n")
		cli.Printf("   Or use --dry-run to see what would be deleted without making changes.\n")
		return nil
	}

	// Perform the actual pruning
	cli.Printf("\n🗑️  Pruning orphaned resources...\n")
	cli.Printf("📍 Application: %s (namespace: %s)\n", myApp.Name, myApp.Namespace)

	// Use ArgoCD's sync with prune option
	prune := true
	dryRun := false

	syncRequest := &application.ApplicationSyncRequest{
		Name:         &myApp.Name,
		AppNamespace: &myApp.Namespace,
		Prune:        &prune,
		DryRun:       &dryRun,
	}

	_, err = apiclient.AppClient.Sync(context.Background(), syncRequest)
	if err != nil {
		return fmt.Errorf("failed to prune orphaned resources: %s", err)
	}

	cli.Printf("✅ Successfully initiated pruning of orphaned resources for application %s\n", myApp.Name)
	cli.Printf("💡 Use 'yak argocd status --application %s' to check the sync status.\n", appName)
	cli.Printf("🔍 Note: If resources still appear orphaned after a successful prune:\n")
	cli.Printf("   - Resources might have finalizers preventing deletion\n")
	cli.Printf("   - Application sync policy might not allow automated pruning\n")
	cli.Printf("   - Resources might be ignored by ArgoCD's orphan detection\n")
	cli.Printf("   Run this command again with --dry-run to see current status.\n")

	return nil
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune orphaned resources for an ArgoCD application",
	Long: `Prune (delete) orphaned resources that are no longer defined in Git but still exist in the cluster.
Orphaned resources are Kubernetes resources that were previously managed by ArgoCD but are no longer 
defined in the application's Git repository.

By default, this command shows what would be pruned without making changes. Use --confirm to actually 
perform the pruning operation.`,
	Example: `yak argocd prune --application my-app --dry-run
yak argocd prune --application my-app --confirm
yak argocd prune --application my-app --json`,
	RunE: prune,
}

func formatResourceIdentifier(resource argocdhelper.AppResource) string {
	identifier := ""
	if resource.Group != "" {
		identifier = resource.Group + "/"
	}
	identifier += resource.Kind
	if resource.Namespace != "" {
		identifier += "/" + resource.Namespace
	}
	identifier += "/" + resource.Name
	return identifier
}

func init() {
	pruneCmd.Flags().StringVarP(&providedPruneFlags.application, "application", "a", "", "ArgoCD application name")
	pruneCmd.Flags().BoolVar(&providedPruneFlags.dryRun, "dry-run", false, "Show what would be pruned without making changes")
	pruneCmd.Flags().BoolVar(&providedPruneFlags.confirm, "confirm", false, "Actually perform the pruning operation")
	pruneCmd.Flags().BoolVar(&providedPruneFlags.verbose, "verbose", false, "Show detailed information about orphaned resources")
	_ = pruneCmd.MarkFlagRequired("application")
}
