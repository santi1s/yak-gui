package tfe

import (
	"context"
	"fmt"

	"github.com/santi1s/yak/internal/constant"
	"github.com/spf13/cobra"
)

type unlockFlags struct {
	organization string
	workspaces   []string
	force        bool
}

func tfeUnlockWorkspace(cmd *cobra.Command, args []string) error {
	config, err := getTfeConfig()
	if err != nil {
		return err
	}

	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, w := range providedUnlockFlags.workspaces {
		// Get the workspace by name
		workspace, err := client.Workspaces.Read(context.Background(), providedUnlockFlags.organization, w)
		if err != nil {
			return fmt.Errorf("error while retrieving workspaces for organization %s: %v", providedUnlockFlags.organization, err)
		}

		if providedUnlockFlags.force {
			_, err = client.Workspaces.ForceUnlock(ctx, workspace.ID)
		} else {
			_, err = client.Workspaces.Unlock(ctx, workspace.ID)
		}
		if err != nil {
			return fmt.Errorf("error while unlocking workspace %s: %v", w, err)
		}
	}
	return nil
}

var (
	providedUnlockFlags   unlockFlags
	tfeWorkspaceUnlockCmd = &cobra.Command{
		Use:   "unlock",
		Short: "unlock workspaces",
		RunE:  tfeUnlockWorkspace,
		Example: "Unlock workspaces ws1,ws2:\n" +
			"unlock --workspaces=ws1,ws2\n\n" +
			"Force unlock workspace ws1:\n" +
			"unlock --force --workspaces=ws1",
	}
)

func init() {
	tfeWorkspaceUnlockCmd.Flags().StringSliceVarP(&providedUnlockFlags.workspaces, "workspaces", "w", []string{}, "workspaces to lock")
	tfeWorkspaceUnlockCmd.Flags().BoolVarP(&providedUnlockFlags.force, "force", "f", false, "force unlock workspace")
	tfeWorkspaceUnlockCmd.Flags().StringVar(&providedUnlockFlags.organization, "organization", constant.TfeDefaultOrganization, "execute command on given organization")
}
