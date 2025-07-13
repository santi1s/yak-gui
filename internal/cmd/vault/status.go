package vault

import (
	"fmt"

	"github.com/spf13/cobra"
)

func status(cmd *cobra.Command, args []string) {
	fmt.Println("status action")
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "print status of a vault cluster",
	Run:   status,
}

func init() {
	//vaultCmd.AddCommand(statusCmd)
}
