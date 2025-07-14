package argocd

import (
	"fmt"
	"net/url"

	"github.com/santi1s/yak/cli"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var providedDashboardFlags ArgoCDDashboardFlags

type ArgoCDDashboardFlags struct {
	application string
}

func dashboard(cmd *cobra.Command, args []string) error {
	// Get ArgoCD server address
	serverAddr, err := argocdhelper.ArgoCDAddr(providedFlags.addr)
	if err != nil {
		return fmt.Errorf("unable to determine ArgoCD server address: %s", err)
	}

	// Build the URL
	var dashboardURL string

	if providedDashboardFlags.application != "" {
		// Deep link to specific application - need to get application details first
		apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
			ArgocdServer: providedFlags.addr,
		})
		if err != nil {
			return fmt.Errorf("unable to establish connection to argocd: %s", err)
		}

		myApp, err := argocdhelper.GetApplication(apiclient.AppClient, providedDashboardFlags.application, providedFlags.project)
		if err != nil {
			return fmt.Errorf("unable to get application %s: %s", providedDashboardFlags.application, err)
		}

		// Use application namespace instead of project name in URL
		dashboardURL = fmt.Sprintf("https://%s/applications/%s/%s",
			serverAddr,
			url.QueryEscape(myApp.Namespace),
			url.QueryEscape(providedDashboardFlags.application))
	} else {
		// General applications view for the project
		dashboardURL = fmt.Sprintf("https://%s/applications?search=%s",
			serverAddr,
			url.QueryEscape(providedFlags.project))
	}

	cli.Printf("Opening ArgoCD dashboard: %s\n", dashboardURL)

	// Open the browser
	err = open.Run(dashboardURL)
	if err != nil {
		return fmt.Errorf("failed to open browser: %s", err)
	}

	if providedDashboardFlags.application != "" {
		cli.Printf("Opened ArgoCD dashboard for application '%s'\n",
			providedDashboardFlags.application)
	} else {
		cli.Printf("Opened ArgoCD dashboard for project '%s'\n", providedFlags.project)
	}

	return nil
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the ArgoCD web dashboard in your browser",
	Example: `yak argocd dashboard
yak argocd dashboard -a my-app`,
	RunE: dashboard,
}

func init() {
	dashboardCmd.Flags().StringVarP(&providedDashboardFlags.application, "application", "a", "", "Open dashboard for specific application")
}
