package tfe

import (
	"context"
	"fmt"
	"strings"

	"github.com/doctolib/yak/cli"
	"github.com/doctolib/yak/internal/constant"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"
)

var (
	organization string
	tfVersion    string
	workspaces   []string
)

func tfeWorkspaceSetVersion(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	config, err := getTfeConfig()
	if err != nil {
		return err
	}

	// Get TFE client from the existing configuration
	client, err := getTfeClient(config)
	if err != nil {
		return err
	}

	successCount := 0
	failedWorkspaces := make([]string, 0)
	skippedWorkspaces := make([]string, 0)

	for _, workspace := range workspaces {
		// Get the workspace
		ws, err := client.Workspaces.Read(ctx, organization, workspace)
		if err != nil {
			cli.Printf("Failed to read workspace %s: %v\n", workspace, err)
			failedWorkspaces = append(failedWorkspaces, workspace)
			continue
		}

		cli.Printf("TF version %s: \n", ws.TerraformVersion)

		// Check if the version is already set
		if ws.TerraformVersion == tfVersion {
			cli.Printf("Skipping workspace %s: already using Terraform version %s\n", workspace, tfVersion)
			skippedWorkspaces = append(skippedWorkspaces, workspace)
			continue
		}

		// Update the workspace with new Terraform version
		options := tfe.WorkspaceUpdateOptions{
			TerraformVersion: tfe.String(tfVersion),
		}

		_, err = client.Workspaces.Update(ctx, organization, workspace, options)
		if err != nil {
			cli.Printf("Failed to update workspace %s: %v\n", workspace, err)
			failedWorkspaces = append(failedWorkspaces, workspace)
			continue
		}

		successCount++
		cli.Printf("Successfully updated Terraform version to %s for workspace: %s\n", tfVersion, workspace)
	}

	cli.Printf("\nSummary:\n")
	cli.Printf("Successfully updated %d workspace(s)\n", successCount)
	if len(skippedWorkspaces) > 0 {
		cli.Printf("Skipped %d workspace(s) (already at target version): %s\n",
			len(skippedWorkspaces),
			strings.Join(skippedWorkspaces, ", "))
	}
	if len(failedWorkspaces) > 0 {
		cli.Printf("Failed to update %d workspace(s): %s\n",
			len(failedWorkspaces),
			strings.Join(failedWorkspaces, ", "))
		return fmt.Errorf("some workspaces failed to update")
	}

	return nil
}

func init() {
	tfeWorkspaceSetVersionCmd.Flags().StringSliceVar(&workspaces, "workspaces", []string{}, "List of workspace names")
	tfeWorkspaceSetVersionCmd.Flags().StringVar(&tfVersion, "version", "", "Terraform version to set")
	tfeWorkspaceSetVersionCmd.Flags().StringVar(&organization, "organization", constant.TfeDefaultOrganization, "TFE Organization")

	err := tfeWorkspaceSetVersionCmd.MarkFlagRequired("workspaces")
	if err != nil {
		panic(err)
	}

	err = tfeWorkspaceSetVersionCmd.MarkFlagRequired("version")
	if err != nil {
		panic(err)
	}

	tfeWorkspaceCmd.AddCommand(tfeWorkspaceSetVersionCmd)
}

var tfeWorkspaceSetVersionCmd = &cobra.Command{
	Use:   "set-version",
	Short: "Set Terraform version for specified workspaces",
	RunE:  tfeWorkspaceSetVersion,
	Example: "Update workspaces MyWs1 and MyWs2 to terraform 1.8.1:\n" +
		"yak tfe workspace set-version --workspaces MyWs1,MyWs2 --version 1.8.1 \n" +
		"\n" +
		"Update workspaces MyWs1 and MyWs2 to terraform 1.8.1 in organization org:\n" +
		"yak tfe workspace set-version --workspaces MyWs1,MyWs2 --version 1.8.1 --organization org",
}
