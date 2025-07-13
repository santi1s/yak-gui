package argocd

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/doctolib/yak/cli"
	argocdhelper "github.com/doctolib/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

type ArgoCDSyncFlags struct {
	application string
	all         bool
	dryRun      bool
	prune       bool
	strategy    string
}

var providedSyncFlags ArgoCDSyncFlags

func syncApp(cmd *cobra.Command, args []string) error {
	appName := providedSyncFlags.application

	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	if providedSyncFlags.all {
		// Get all applications in the project
		myApps, err := argocdhelper.GetAllApplications(apiclient.AppClient, providedFlags.project)
		if err != nil {
			return fmt.Errorf("unable to get applications: %s", err)
		}

		for _, app := range myApps.Items {
			err := syncApplication(apiclient.AppClient, app.Name, app.Namespace)
			if err != nil {
				cli.Printf("Failed to sync application %s: %s\n", app.Name, err)
			} else {
				cli.Printf("Successfully triggered sync for application %s\n", app.Name)
			}
		}
	} else {
		// Sync single application
		myApp, err := argocdhelper.GetApplication(apiclient.AppClient, appName, providedFlags.project)
		if err != nil {
			return err
		}

		err = syncApplication(apiclient.AppClient, myApp.Name, myApp.Namespace)
		if err != nil {
			return fmt.Errorf("failed to sync application %s: %s", appName, err)
		}
		cli.Printf("Successfully triggered sync for application %s\n", appName)
	}

	return nil
}

func syncApplication(appClient application.ApplicationServiceClient, appName, appNamespace string) error {
	syncRequest := &application.ApplicationSyncRequest{
		Name:         &appName,
		AppNamespace: &appNamespace,
		DryRun:       &providedSyncFlags.dryRun,
		Prune:        &providedSyncFlags.prune,
	}

	// Set sync strategy if provided
	if providedSyncFlags.strategy != "" {
		var strategy *v1alpha1.SyncStrategy
		switch providedSyncFlags.strategy {
		case "apply":
			strategy = &v1alpha1.SyncStrategy{
				Apply: &v1alpha1.SyncStrategyApply{
					Force: false,
				},
			}
		case "hook":
			strategy = &v1alpha1.SyncStrategy{
				Hook: &v1alpha1.SyncStrategyHook{},
			}
		default:
			return fmt.Errorf("invalid sync strategy: %s. Valid options are 'apply' or 'hook'", providedSyncFlags.strategy)
		}
		syncRequest.Strategy = strategy
	}

	_, err := appClient.Sync(context.Background(), syncRequest)
	return err
}

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Sync ArgoCD applications",
	Example: "yak argocd sync --application my-app\nyak argocd sync --all",
	RunE:    syncApp,
}

func init() {
	syncCmd.Flags().StringVarP(&providedSyncFlags.application, "application", "a", "", "ArgoCD application name")
	syncCmd.Flags().BoolVar(&providedSyncFlags.all, "all", false, "Sync all applications")
	syncCmd.Flags().BoolVar(&providedSyncFlags.dryRun, "dry-run", false, "Preview sync without making changes")
	syncCmd.Flags().BoolVar(&providedSyncFlags.prune, "prune", false, "Allow deleting unexpected resources")
	syncCmd.Flags().StringVar(&providedSyncFlags.strategy, "strategy", "", "Sync strategy: 'apply' (kubectl apply) or 'hook' (use hooks)")
	syncCmd.MarkFlagsMutuallyExclusive("application", "all")
	syncCmd.MarkFlagsOneRequired("application", "all")
}
