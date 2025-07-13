package tfe

import (
	"context"
	"fmt"

	"github.com/doctolib/yak/internal/constant"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

type lockFlags struct {
	organization string
	workspaces   []string
	checkStatus  bool
}

func tfeLockWorkspace(cmd *cobra.Command, args []string) error {
	config, err := getTfeConfig()
	if err != nil {
		return err
	}

	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	options := &tfe.WorkspaceLockOptions{Reason: tfe.String("Locking workspace for maintenance")}

	for _, w := range providedLockFlags.workspaces {
		// Get the workspace by name
		workspace, err := client.Workspaces.Read(context.Background(), providedLockFlags.organization, w)
		if err != nil {
			return fmt.Errorf("error while retrieving workspaces for organization %s: %v", providedLockFlags.organization, err)
		}

		if providedLockFlags.checkStatus {
			run, err := client.Runs.List(ctx, workspace.ID, &tfe.RunListOptions{
				Operation:   string(tfe.RunOperationPlanApply),
				ListOptions: tfe.ListOptions{PageSize: 1},
			})

			if err != nil {
				return fmt.Errorf("error listing runs for workspace: %w", err)
			}

			if len(run.Items) > 0 {
				latestRun := run.Items[0]
				if latestRun.Status != tfe.RunApplied && (latestRun.HasChanges && latestRun.Status == tfe.RunPlannedAndFinished) {
					return fmt.Errorf("latest run for workspace %s is not applied, current status: %s", w, latestRun.Status)
				}
			}
		}
		_, err = client.Workspaces.Lock(ctx, workspace.ID, *options)
		if err != nil {
			return fmt.Errorf("error while locking workspace %s: %v", w, err)
		}
	}
	return nil
}

var (
	providedLockFlags   lockFlags
	tfeWorkspaceLockCmd = &cobra.Command{
		Use:   "lock",
		Short: "lock workspaces",
		RunE:  tfeLockWorkspace,
		Example: "Lock workspaces ws1,ws2:\n" +
			"lock --check-status --workspaces=ws1,ws2\n\n" +
			"Lock workspace ws1 only if last apply status is applied:\n" +
			"lock --check-status --workspaces=ws1",
	}
)

func init() {
	tfeWorkspaceLockCmd.Flags().StringSliceVarP(&providedLockFlags.workspaces, "workspaces", "w", []string{}, "workspaces to lock")
	tfeWorkspaceLockCmd.Flags().BoolVarP(&providedLockFlags.checkStatus, "check-status", "s", false, "check workspace status")
	tfeWorkspaceLockCmd.Flags().StringVar(&providedLockFlags.organization, "organization", constant.TfeDefaultOrganization, "execute command on given organization")
}
