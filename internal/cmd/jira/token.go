package jira

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type tasksJiraTokenFlags struct {
	token string
}

var (
	providedJiraTokenFlags tasksJiraTokenFlags
	jiraTokenCmd           = &cobra.Command{
		Use:     "token",
		Aliases: []string{"jt"},
		Short:   "Configure Jira API token",
		RunE:    jiraToken,
		Example: "yak jira jt -t <Jira API token>",
	}
)

// jiraToken saves the provided Jira API token to the system keyring
func jiraToken(_ *cobra.Command, _ []string) error {
	// Get the user's email from the global Git config
	account, err := GetGitConfigEmail(&GitConfig{})
	if err != nil {
		log.Errorf("Error getting user email from the global Git config: %v", err)
		return err
	}

	// Save the Jira API token to the keyring
	// The keyring uses the user's email as a unique identifier
	if err := keyring.Set(service, account, providedJiraTokenFlags.token); err != nil {
		return err
	}

	// Everything went well, return no error
	return nil
}

func init() {
	jiraTokenCmd.Flags().StringVarP(&providedJiraTokenFlags.token, "token", "t", "", "Provide JIRA API Token - generate token in https://id.atlassian.com/manage-profile/security/api-tokens")
	err := jiraTokenCmd.MarkFlagRequired("token")
	if err != nil {
		log.Errorf("Error getting user email from the global Git config: %v", err)
	}
}
