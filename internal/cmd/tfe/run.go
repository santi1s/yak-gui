package tfe

import (
	"github.com/spf13/cobra"
)

type runFlags struct {
	organization   string
	dryrun         bool
	age            int
	discardPending bool
	allWorkspaces  bool
	workspaces     []string
}

var (
	providedRunFlags runFlags
	tfeRunCmd        = &cobra.Command{
		Use:   "run",
		Short: "manage TFE runs",
	}
)

func init() {
	tfeRunCmd.AddCommand(tfeRunDiscardCmd)
}
