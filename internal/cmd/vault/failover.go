package vault

import (
	"fmt"

	"github.com/spf13/cobra"
)

func failover(cmd *cobra.Command, args []string) {
	fmt.Println("failover action")
}

var failoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "failover a vault cluster",
	Run:   failover,
}

func init() {
	//vaultCmd.AddCommand(failoverCmd)
}
