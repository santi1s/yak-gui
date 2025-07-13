package vault

import (
	"fmt"

	"github.com/spf13/cobra"
)

func setupReplication(cmd *cobra.Command, args []string) {
	fmt.Println("setup-replication action")
}

var setupreplicationCmd = &cobra.Command{
	Use:   "setup-replication",
	Short: "setup replication between two vault clusters",
	Run:   setupReplication,
}

func init() {
	//vaultCmd.AddCommand(setupreplicationCmd)
}
