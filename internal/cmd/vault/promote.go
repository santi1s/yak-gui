package vault

import (
	"fmt"

	"github.com/spf13/cobra"
)

func promote(cmd *cobra.Command, args []string) {
	fmt.Println("promote action")
}

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "promote a vault cluster in case of disaster",
	Run:   promote,
}

func init() {
	//vaultCmd.AddCommand(promoteCmd)
}
