package tfe

import (
	"context"
	"fmt"
	"time"

	"github.com/santi1s/yak/cli"
	"github.com/santi1s/yak/internal/constant"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

var tm = time.Now().UTC()
var discardComment = "Run automatically discarded because the plan was older than set age"
var cancelComment = "Run automatically canceled because the plan was older than set age"

func formatRunLink(wsName, runID string) string {
	return fmt.Sprintf("https://tfe.doctolib.net/app/doctolib/workspaces/%s/runs/%s", wsName, runID)
}

func computeWorkspaces(ctx context.Context, client *tfe.Client, workspaces []string) (*tfe.WorkspaceList, error) {
	computedWorkspaces := &tfe.WorkspaceList{}
	for _, workspace := range workspaces {
		ws, err := client.Workspaces.Read(ctx, "doctolib", workspace)
		if err != nil {
			return nil, err
		}
		computedWorkspaces.Items = append(computedWorkspaces.Items, ws)
	}
	return computedWorkspaces, nil
}

func tfeDiscardRuns(_ *cobra.Command, args []string) error {
	var wsList = new(tfe.WorkspaceList)
	config, err := getTfeConfig()
	if err != nil {
		return err
	}
	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	if providedRunFlags.dryrun {
		fmt.Println("dry-run mode, not actually doing anything.")
	}

	ctx := context.Background()
	maxRunAge, err := time.ParseDuration(fmt.Sprint(providedRunFlags.age, "h0m0s"))
	if err != nil {
		return err
	}
	discardedRuns := []string{}
	runOptions := &tfe.RunListOptions{}
	page := 1
	for {
		wsOptions := &tfe.WorkspaceListOptions{
			ListOptions: tfe.ListOptions{
				PageSize:   20,
				PageNumber: page,
			},
		}

		if providedRunFlags.discardPending {
			runOptions.Status = "planned,pending"
		} else {
			runOptions.Status = "planned"
		}

		discardRunOptions := tfe.RunDiscardOptions{
			Comment: &discardComment,
		}

		cancelRunOptions := tfe.RunCancelOptions{
			Comment: &discardComment,
		}

		if len(providedRunFlags.workspaces) > 0 {
			wsList, err = computeWorkspaces(ctx, client, providedRunFlags.workspaces)
		} else {
			wsList, err = client.Workspaces.List(ctx, "doctolib", wsOptions)
		}
		if err != nil {
			return fmt.Errorf("error while listing workspaces for organization %s: %v", providedRunFlags.organization, err)
		}

		for _, w := range wsList.Items {
			runs, err := client.Runs.List(ctx, w.ID, runOptions)
			if err != nil {
				return fmt.Errorf("error while listing runs for workspace %s (ID %s): %v", w.Name, w.ID, err)
			}
			if len(runs.Items) > 0 {
				for _, r := range runs.Items {
					runCreationTime := r.CreatedAt
					if tm.Sub(runCreationTime) > maxRunAge {
						if providedRunFlags.dryrun {
							fmt.Printf("dry-run -- would've discarded/canceled run with ID %s for workspace %s\n", r.ID, w.Name)
						} else {
							if r.Status == "pending" {
								err = client.Runs.Cancel(ctx, r.ID, cancelRunOptions)
							} else {
								err = client.Runs.Discard(ctx, r.ID, discardRunOptions)
							}
							if err != nil {
								return fmt.Errorf("error while discarding run %s for workspace %s", r.ID, w.ID)
							}
							discardedRuns = append(discardedRuns, formatRunLink(w.Name, r.ID))
						}
					}
				}
			}
		}

		if providedRunFlags.allWorkspaces {
			if wsList.CurrentPage == wsList.TotalPages || wsList.TotalPages == 0 {
				break
			}
		} else {
			break
		}

		page++
	}

	return cli.PrintYAML(discardedRuns)
}

var (
	tfeRunDiscardCmd = &cobra.Command{
		Use:   "discard",
		Short: "discard old runs",
		RunE:  tfeDiscardRuns,
	}
)

func init() {
	tfeRunDiscardCmd.PersistentFlags().BoolVar(&providedRunFlags.discardPending, "discard-pending", false, "discard run in pending state too")
	tfeRunDiscardCmd.PersistentFlags().BoolVar(&providedRunFlags.dryrun, "dry-run", false, "enables dry run")
	tfeRunDiscardCmd.PersistentFlags().StringVar(&providedRunFlags.organization, "organization", constant.TfeDefaultOrganization, "execute command on given organization")
	tfeRunDiscardCmd.PersistentFlags().IntVar(&providedRunFlags.age, "age", 12, "runs older than $age hours will be discarded")
	tfeRunDiscardCmd.PersistentFlags().BoolVar(&providedRunFlags.allWorkspaces, "all-workspaces", false, "discard pending runs in all workspaces")
	tfeRunDiscardCmd.PersistentFlags().StringSliceVar(&providedRunFlags.workspaces, "workspaces", []string{}, "discard pending runs in selected workspaces")
	tfeRunDiscardCmd.MarkFlagsMutuallyExclusive("all-workspaces", "workspaces")
}
