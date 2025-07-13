package argocd

import (
	"fmt"
	"sort"

	"github.com/doctolib/yak/cli"
	argocdhelper "github.com/doctolib/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

type ArgoCDOrphanFlags struct {
	application   string
	nonNamespaced bool
}

var providedOrphanFlags ArgoCDOrphanFlags

func orphaned(cmd *cobra.Command, args []string) error {
	projectName := providedFlags.project
	appName := providedOrphanFlags.application
	nonNamespaced := providedOrphanFlags.nonNamespaced

	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	_, err = argocdhelper.GetApplication(apiclient.AppClient, appName, projectName)
	if err != nil {
		return err
	}

	orphanedResults := make(map[string][]argocdhelper.AppResource)
	orphanedResultsFiltered := make(map[string][]argocdhelper.AppResource)

	// if we didn't specify nonNamespaced only, we need application orphans resources
	if !nonNamespaced {
		// orphaned Namespaced Resources in all apps
		orphanedResourcesByApp, err := argocdhelper.OrphanedResourcesArgoCD(apiclient.AppClient, projectName)
		if err != nil {
			return fmt.Errorf("unable to get orphaned resources: %s", err)
		}
		orphanedResults = orphanedResourcesByApp
	}
	// if application is not specified, we need non-namespaced apps
	if appName == "" {
		// Get orphaned non-namespaced Resources
		remainingOrphanedResources, err := argocdhelper.GetOrphanedNonNamespacedResources(apiclient.AppClient, projectName)
		if err != nil {
			return fmt.Errorf("unable to get non-namespaced orphaned resources: %s", err)
		}
		orphanedResults[""] = remainingOrphanedResources
	}

	if nonNamespaced {
		// filter on non-namespaced only
		orphanedResultsFiltered[""] = orphanedResults[""]
	} else if appName != "" {
		orphanedResultsFiltered[appName] = orphanedResults[appName]
	} else {
		// no filtering, keep everything
		orphanedResultsFiltered = orphanedResults
	}

	if providedFlags.json || providedFlags.yaml {
		err := formatOutput(orphanedResultsFiltered)
		if err != nil {
			return err
		}
	} else {
		formatOutputOrphanedResources(orphanedResultsFiltered)
	}
	return nil
}

func formatOutputOrphanedResources(orphans map[string][]argocdhelper.AppResource) {
	const appNameWidth = 20
	const orphanResourceNameWidth = 60
	const orphanResourceTypeWidth = 30
	cli.Println("** Orphaned Resources **")
	cli.Printf("%-*s %-*s %-*s %s\n", appNameWidth, "Application", orphanResourceNameWidth, "Name", orphanResourceTypeWidth, "Kind", "Group")

	var apps []string
	for app := range orphans {
		apps = append(apps, app)
	}
	sort.Strings(apps)

	for _, app := range apps {
		orphanItems := orphans[app]
		for _, item := range orphanItems {
			cli.Printf("%-*s %-*s %-*s %s\n", appNameWidth, app, orphanResourceNameWidth, item.Name, orphanResourceTypeWidth, item.Kind, item.Group)
		}
	}
}

var orphanResourcesCmd = &cobra.Command{
	Use:     "orphaned",
	Short:   "Get the orphan resources in argocd",
	Example: `yak argocd orphaned [-a my-app] [-N]`,
	RunE:    orphaned,
}

func init() {
	orphanResourcesCmd.Flags().StringVarP(&providedOrphanFlags.application, "application", "a", "", "filter on application name")
	orphanResourcesCmd.Flags().BoolVarP(&providedOrphanFlags.nonNamespaced, "non-namespaced", "N", false, "set this to true to display only the non-namespaced orphan resources (cluster wide)")
	orphanResourcesCmd.MarkFlagsMutuallyExclusive("application", "non-namespaced")
}
