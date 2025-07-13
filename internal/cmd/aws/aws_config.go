package aws

import (
	"github.com/spf13/cobra"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "AWS config management",
	}
)

func init() {
	configCmd.AddCommand(requestConfigCmd)
	configCmd.AddCommand(generateConfigCmd)
}
