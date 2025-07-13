package vault

import (
	"fmt"

	"github.com/spf13/cobra"
)

func removePeer(cmd *cobra.Command, args []string) {
	fmt.Println("remove-peer action")
}

var removePeerCmd = &cobra.Command{
	Use:   "remove-peer",
	Short: "remove peer from a vault cluster",
	Run:   removePeer,
}

func init() {
	//vaultCmd.AddCommand(removePeerCmd)
}
