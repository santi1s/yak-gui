package argocd

import (
	"fmt"

	"github.com/santi1s/yak/cli"
	argocdhelper "github.com/santi1s/yak/internal/helper/argocd"
	"github.com/spf13/cobra"
)

var providedUnsuspendFlags ArgoCDUnsuspendFlags

type ArgoCDUnsuspendFlags struct {
	application string
	all         bool
}

func unsuspend(cmd *cobra.Command, args []string) error {
	appName := providedUnsuspendFlags.application

	// Authentication
	apiclient, err := argocdhelper.ArgocdLogin(&argocdhelper.LoginParams{
		ArgocdServer: providedFlags.addr,
	})
	if err != nil {
		return fmt.Errorf("unable to establish connection to argocd: %s", err)
	}

	myProject, err := argocdhelper.GetArgoCDProject(apiclient.ProjectClient, providedFlags.project)
	if err != nil {
		return err
	}

	if providedUnsuspendFlags.all {
		// Update the window
		err := argocdhelper.UpdateWindow(myProject, []string{})
		if err != nil {
			return err
		}

		err = argocdhelper.UpdateArgoCDProject(apiclient.ProjectClient, myProject)
		if err != nil {
			return err
		}
		cli.Print("All the applications are now unsuspended\n")
	} else {
		myApp, err := argocdhelper.GetApplication(apiclient.AppClient, appName, myProject.Name)
		if err != nil {
			return err
		}

		// Empty windows list
		windowsList := myProject.Spec.SyncWindows
		if windowsList == nil {
			cli.Printf("The application %s is already active for sync\n", appName)
		} else {
			// Get the window for this application
			w := windowsList.Matches(myApp)
			if w == nil {
				cli.Printf("The application %s is already active for sync\n", appName)
			} else {
				for _, window := range windowsList {
					newApps := []string{}
					for _, app := range window.Applications {
						if app != appName {
							newApps = append(newApps, app)
						}
					}
					window.Applications = newApps
				}

				err = argocdhelper.UpdateArgoCDProject(apiclient.ProjectClient, myProject)
				if err != nil {
					return err
				}
				cli.Printf("The application %s is now unsuspended\n", appName)
			}
		}
	}
	return nil
}

var unsuspendCmd = &cobra.Command{
	Use:     "unsuspend",
	Short:   "Unsuspend an ArgoCD Application.",
	Example: "yak argocd unsuspend --application my-app\nyak argocd unsuspend --all",
	RunE:    unsuspend,
}

func init() {
	unsuspendCmd.Flags().StringVarP(&providedUnsuspendFlags.application, "application", "a", "", "ArgoCD application name")
	unsuspendCmd.Flags().BoolVar(&providedUnsuspendFlags.all, "all", false, "Unsuspend all applications")
	unsuspendCmd.MarkFlagsMutuallyExclusive("application", "all")
	unsuspendCmd.MarkFlagsOneRequired("application", "all")
}
