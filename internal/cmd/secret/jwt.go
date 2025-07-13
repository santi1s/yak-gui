package secret

import (
	"github.com/spf13/cobra"
)

var (
	jwtCmd = &cobra.Command{
		Use:   "jwt",
		Short: "Manage creation of interservice communication JWT token secret in vault",
		Long:  "Manage creation of interservice communication JWT token secret in vault\nMore info:\nhttps://doctolib.atlassian.net/wiki/spaces/TTP/pages/1578042365/Understanding+JWT+config",
	}
)

func init() {
	jwtCmd.AddCommand(jwtClientCmd)
	jwtCmd.AddCommand(jwtServerCmd)
	jwtCmd.AddCommand(jwtLintCmd)
}
