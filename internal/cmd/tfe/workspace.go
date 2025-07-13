package tfe

import (
	"github.com/spf13/cobra"
)

var (
	tfeWorkspaceCmd = &cobra.Command{
		Use:   "workspace",
		Short: "manage workspace resources",
	}
)

func init() {
	tfeWorkspaceCmd.AddCommand(tfeWorkspaceListCmd)
	tfeWorkspaceCmd.AddCommand(tfeWorkspaceLockCmd)
	tfeWorkspaceCmd.AddCommand(tfeWorkspaceUnlockCmd)
}
