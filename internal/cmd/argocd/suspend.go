package argocd

import (
	"fmt"

	"github.com/doctolib/yak/cli"
	argocdhelper "github.com/doctolib/yak/internal/helper/argocd"

	"github.com/spf13/cobra"
)

var providedSuspendFlags ArgoCDSuspendFlags

type ArgoCDSuspendFlags struct {
	application string
	all         bool
}

func suspend(cmd *cobra.Command, args []string) error {
	appName := providedSuspendFlags.application

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
		return err
	}

	// Get the windows list
	windowsList := myProject.Spec.SyncWindows
	newApps := []string{}

	// If --all flag is set, suspend all applications
	if providedSuspendFlags.all {
		// Get all applications
		myApps, _ := argocdhelper.GetAllApplications(apiclient.AppClient, myProject.Name)

		for _, app := range myApps.Items {
			newApps = append(newApps, app.Name)
		}
		// Update the window
		err := argocdhelper.UpdateWindow(myProject, newApps)
		if err != nil {
			return err
		}
	} else {
		// Suspend the application provided in the flag
		myApp, err := argocdhelper.GetApplication(apiclient.AppClient, appName, myProject.Name)
		if err != nil {
			return err
		}

		// Empty windows list
		if windowsList == nil {
			// Need to add the first window with the target application
			newApps = append(newApps, appName)
			err := argocdhelper.UpdateWindow(myProject, newApps)
			if err != nil {
				return err
			}
		} else {
			// Get the window for this application
			w := windowsList.Matches(myApp)
			if w != nil {
				cli.Printf("Application %s already suspended\n", appName)
			}

			windowIdx := argocdhelper.GetSyncWindow(myProject)
			// Add the application to the window, along with already existing applications
			newApps := append(windowsList[windowIdx].Applications, appName)

			// Update the window
			err := argocdhelper.UpdateWindow(myProject, newApps)
			if err != nil {
				return err
			}
		}
	}
	// Update the project
	err = argocdhelper.UpdateArgoCDProject(apiclient.ProjectClient, myProject)
	if err != nil {
		return err
	}
	if providedSuspendFlags.all {
		cli.Printf("All applications are now suspended\n")
	} else {
		cli.Printf("The application [%s] is now suspended\n", appName)
	}
	return nil
}

var suspendCmd = &cobra.Command{
	Use:     "suspend",
	Short:   "Suspend an ArgoCD Application.",
	Example: "yak argocd suspend --application my-app\nyak argocd suspend --all",
	RunE:    suspend,
}

func init() {
	suspendCmd.Flags().StringVarP(&providedSuspendFlags.application, "application", "a", "", "ArgoCD application name")
	suspendCmd.Flags().BoolVar(&providedSuspendFlags.all, "all", false, "Suspend all applications")
	suspendCmd.MarkFlagsMutuallyExclusive("application", "all")
	suspendCmd.MarkFlagsOneRequired("application", "all")
}
