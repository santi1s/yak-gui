package jira

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	taskCmd = &cobra.Command{
		Use:   "task",
		Short: "Jira Task related commands",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			account, err := GetGitConfigEmail(&GitConfig{})
			if err == nil {
				_, err = GetJiraClient(account, providedFlags.jiraURL)
				if err != nil {
					fmt.Printf("Secret 'Jira API token' not found in keyring. Generate one with 'yak jira token'\n")
					log.Exit(1)
				}
			}
			return nil
		},
	}
)

func init() {
	taskCmd.AddCommand(listTasksCmd)
	taskCmd.AddCommand(createBranchCmd)
	taskCmd.AddCommand(createCommitCmd)
}
