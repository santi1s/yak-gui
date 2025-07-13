package argocd

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/doctolib/yak/cli"
	argocdhelper "github.com/doctolib/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

type ArgoCDRefreshFlags struct {
	application string
	all         bool
	hard        bool
}

var providedRefreshFlags ArgoCDRefreshFlags

func refreshApp(cmd *cobra.Command, args []string) error {
	appName := providedRefreshFlags.application

	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	if providedRefreshFlags.all {
		// Get all applications in the project
		myApps, err := argocdhelper.GetAllApplications(apiclient.AppClient, providedFlags.project)
		if err != nil {
			return fmt.Errorf("unable to get applications: %s", err)
		}

		for _, app := range myApps.Items {
			err := refreshApplication(apiclient.AppClient, app.Name, app.Namespace)
			if err != nil {
				cli.Printf("Failed to refresh application %s: %s\n", app.Name, err)
			} else {
				cli.Printf("Successfully triggered refresh for application %s\n", app.Name)
			}
		}
	} else {
		// Refresh single application
		myApp, err := argocdhelper.GetApplication(apiclient.AppClient, appName, providedFlags.project)
		if err != nil {
			return err
		}

		err = refreshApplication(apiclient.AppClient, myApp.Name, myApp.Namespace)
		if err != nil {
			return fmt.Errorf("failed to refresh application %s: %s", appName, err)
		}
		cli.Printf("Successfully triggered refresh for application %s\n", appName)
	}

	return nil
}

func refreshApplication(appClient application.ApplicationServiceClient, appName, appNamespace string) error {
	refreshType := "normal"
	if providedRefreshFlags.hard {
		refreshType = "hard"
	}

	// ArgoCD refresh is done by calling Get with refresh parameter
	_, err := appClient.Get(context.Background(), &application.ApplicationQuery{
		Name:         &appName,
		AppNamespace: &appNamespace,
		Refresh:      &refreshType,
	})

	return err
}

var refreshCmd = &cobra.Command{
	Use:     "refresh",
	Short:   "Refresh ArgoCD applications",
	Long:    "Refresh ArgoCD applications to detect changes in Git repository",
	Example: "yak argocd refresh --application my-app\nyak argocd refresh --all\nyak argocd refresh --application my-app --hard",
	RunE:    refreshApp,
}

func init() {
	refreshCmd.Flags().StringVarP(&providedRefreshFlags.application, "application", "a", "", "ArgoCD application name")
	refreshCmd.Flags().BoolVar(&providedRefreshFlags.all, "all", false, "Refresh all applications")
	refreshCmd.Flags().BoolVar(&providedRefreshFlags.hard, "hard", false, "Perform hard refresh (clear cache)")
	refreshCmd.MarkFlagsMutuallyExclusive("application", "all")
	refreshCmd.MarkFlagsOneRequired("application", "all")
}
