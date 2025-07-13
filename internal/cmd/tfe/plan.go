package tfe

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/constant"
	"github.com/hashicorp/go-tfe"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type RunInfo struct {
	URL    *string       `json:"url"`
	Error  error         `json:"error"`
	Status tfe.RunStatus `json:"status,omitempty"`
	Result *PlanResult   `json:"result,omitempty"`
	ID     string        `json:"-"`
}

type PlanResult struct {
	HasChanges   bool `json:"haschanges"`
	Additions    int  `json:"additions"`
	Changes      int  `json:"changes"`
	Destructions int  `json:"destructions"`
}

var errTerraformVersionsList = errors.New("error while getting terraform versions list")
var errTerraformVersionIsInvalid = errors.New("terraform version is not a valid semver version")
var errPlanMissingFlags = errors.New("you must provide --workspaces or --owner flag")
var errWorkspaceNotFound = errors.New("workspace not found")

func tfePlan(cmd *cobra.Command, args []string) error {
	if len(providedPlanFlags.workspaces) == 0 && providedPlanFlags.owner == "" {
		return errPlanMissingFlags
	}

	config, err := getTfeConfig()
	if err != nil {
		return err
	}
	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if !semver.IsValid(fmt.Sprintf("v%s", providedPlanFlags.version)) {
		return errTerraformVersionIsInvalid
	}

	workspacesToCreateARun := []*tfe.Workspace{}
	if len(providedPlanFlags.workspaces) > 0 { // workspaces have been provided
		for _, v := range providedPlanFlags.workspaces {
			ws, err := client.Workspaces.Read(ctx, providedPlanFlags.organization, v)
			if err != nil {
				return fmt.Errorf("error while reading workspace %s: %v", v, err)
			}
			workspacesToCreateARun = append(workspacesToCreateARun, ws)
		}
	} else { // owner has been provided
		page := 1
		for {
			wsList, err := client.Workspaces.List(ctx, providedPlanFlags.organization, &tfe.WorkspaceListOptions{
				Tags: fmt.Sprintf("owner:%s", providedPlanFlags.owner),
				ListOptions: tfe.ListOptions{
					PageNumber: page,
				},
			})

			if err != nil {
				return fmt.Errorf("error while listing workspaces on organization %s for owner %s: %v", providedPlanFlags.organization, providedPlanFlags.owner, err)
			}

			workspacesToCreateARun = append(workspacesToCreateARun, wsList.Items...)
			page++

			if wsList.CurrentPage == wsList.TotalPages || wsList.TotalPages == 0 {
				break
			}
		}
	}

	urls := map[string]*RunInfo{}
	var wg sync.WaitGroup
	wg.Add(len(workspacesToCreateARun))

	var bar *progressbar.ProgressBar
	if providedPlanFlags.wait {
		bar = progressbar.Default(int64(len(workspacesToCreateARun) * 2))
	} else {
		bar = progressbar.Default(int64(len(workspacesToCreateARun)))
	}

	for _, workspace := range workspacesToCreateARun {
		go func(ws *tfe.Workspace) {
			defer wg.Done()
			tfeRunInfo := createTerraformPlanOnlyRun(client, ws, providedPlanFlags.version)
			_ = bar.Add(1)

			if providedPlanFlags.wait {
				pr := &PlanResult{}
				pr.Additions = -1
				pr.Changes = -1
				pr.Destructions = -1

				if tfeRunInfo.Error != nil {
					tfeRunInfo.Status = "error while creating tfe run"
					tfeRunInfo.Result = pr
				} else {
					for {
						errorCount := 0
						run, err := client.Runs.Read(ctx, tfeRunInfo.ID)
						if err != nil {
							errorCount++
							time.Sleep(time.Second * 5)
						}

						if errorCount >= 5 {
							tfeRunInfo.Status = "too many errors while querying tfe for run status"
							tfeRunInfo.Result = pr
							break
						}

						if run.Status == "policy_soft_failed" || run.Status == "planned_and_finished" || run.Status == "discarded" || run.Status == "errored" || run.Status == "canceled" || run.Status == "force_canceled" {
							tfeRunInfo.Status = run.Status
							pr.Additions = run.Plan.ResourceAdditions
							pr.Changes = run.Plan.ResourceChanges
							pr.Destructions = run.Plan.ResourceDestructions
							pr.HasChanges = run.Plan.HasChanges
							tfeRunInfo.Result = pr
							break
						}

						time.Sleep(time.Second * 3)
					}
				}

				_ = bar.Add(1)
			}

			urls[ws.Name] = tfeRunInfo
		}(workspace)
	}
	wg.Wait()

	switch {
	case providedPlanFlags.json:
		return cli.PrintJSON(urls)
	case providedPlanFlags.yaml:
		return cli.PrintYAML(urls)
	default:
		return cli.PrintYAML(urls)
	}
}

func createTerraformPlanOnlyRun(client *tfe.Client, workspace *tfe.Workspace, version string) *RunInfo {
	run, err := client.Runs.Create(context.Background(), tfe.RunCreateOptions{
		Message:          tfe.String(fmt.Sprintf("run triggered by yak tfe plan command - tf version %s", version)),
		PlanOnly:         tfe.Bool(true),
		TerraformVersion: tfe.String(version),
		Workspace:        workspace,
	})

	tfeRunInfo := &RunInfo{}

	if err == nil {
		tfeRunInfo.URL = tfe.String(fmt.Sprintf("https://%s/app/%s/workspaces/%s/runs/%s", tfeDefaultHostname, providedPlanFlags.organization, workspace.Name, run.ID))
		tfeRunInfo.ID = run.ID
	} else {
		tfeRunInfo.Error = err
	}

	return tfeRunInfo
}

type planFlags struct {
	json         bool
	owner        string
	version      string
	wait         bool
	workspaces   []string
	yaml         bool
	organization string
}

const tfeDefaultHostname = "tfe.doctolib.net"

var (
	providedPlanFlags planFlags
	tfePlanCmd        = &cobra.Command{
		Use:   "plan",
		Short: "execute a speculative plan on a workspace",
		RunE:  tfePlan,
	}
)

func init() {
	tfePlanCmd.Flags().StringVar(&providedPlanFlags.organization, "organization", constant.TfeDefaultOrganization, "execute command on given organization")
	tfePlanCmd.Flags().StringVarP(&providedPlanFlags.owner, "owner", "o", "", "execute a plan on all workspaces of the given owner")
	tfePlanCmd.Flags().StringVarP(&providedPlanFlags.version, "version", "v", constant.TerraformDefaultVersion, "terraform version to use")
	tfePlanCmd.Flags().StringSliceVarP(&providedPlanFlags.workspaces, "workspaces", "w", []string{}, "execute a plan on a list of workspaces")
	tfePlanCmd.Flags().BoolVar(&providedPlanFlags.wait, "wait", false, "wait for execution of all plans and show result")
	tfePlanCmd.PersistentFlags().BoolVar(&providedPlanFlags.json, "json", false, "format output in JSON")
	tfePlanCmd.PersistentFlags().BoolVar(&providedPlanFlags.yaml, "yaml", false, "format output in YAML")
	tfePlanCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	err := tfePlanCmd.MarkFlagRequired("version")
	if err != nil {
		panic(err)
	}
	tfePlanCmd.MarkFlagsMutuallyExclusive("owner", "workspaces")
}
