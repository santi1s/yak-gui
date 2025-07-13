package secret

import (
	"github.com/spf13/cobra"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "manage a secret metadatas in vault",
}

func init() {
	metadataCmd.AddCommand(metadataGetCmd)
	metadataCmd.AddCommand(metadataCreateCmd)
	metadataCmd.AddCommand(metadataDeleteCmd)
	metadataCmd.AddCommand(metadataUpdateCmd)
}
