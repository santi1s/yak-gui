package jira

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	jira "github.com/andygrunwald/go-jira"
	"github.com/santi1s/yak/internal/constant"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type jiraFlags struct {
	jiraURL string
}

var (
	providedFlags jiraFlags

	jiraCmd = &cobra.Command{
		Use:   "jira",
		Short: "Jira Task related commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Root().Name() == constant.CliName && cmd.Root().PersistentPreRun != nil {
				cmd.Root().PersistentPreRun(cmd, args)
			}

			if cmd.HasParent() && cmd.Parent().Name() == "completion" {
				switch cmd.Name() {
				case "bash", "zsh", "fish", "powershell":
					cmd.ResetFlags()
				}
			}
			cmd.SilenceUsage = true
		},
	}
)

func GetRootCmd() *cobra.Command {
	return jiraCmd
}

type GitConfigInterface interface {
	LoadConfig(scope config.Scope) (*config.Config, error)
}

type GitConfig struct{}

func (g *GitConfig) LoadConfig(scope config.Scope) (*config.Config, error) {
	// Implement the logic to load the Git configuration here
	return config.LoadConfig(config.GlobalScope)
}

// GetGitConfigEmail returns the email from the global Git config
func GetGitConfigEmail(gitConfig GitConfigInterface) (string, error) {
	// Load the global Git config
	// cfg, err := config.LoadConfig(config.GlobalScope)
	cfg, err := GitConfigInterface.LoadConfig(gitConfig, config.GlobalScope)
	if err != nil {
		log.Printf("Error loading global Git config: %v", err)
		return "", err
	}
	// Get the email from the config
	email := cfg.User.Email

	// Check if the email is empty
	if email == "" {
		// Return an error if the email is not found
		return email, fmt.Errorf("email not found")
	}
	// Return the email
	return email, nil
}

// GetJiraClient returns a Jira client with the user's API token
func GetJiraClient(account string, jiraURL string) (*jira.Client, error) {
	// Create a new Jira client
	var client *jira.Client

	// Retrieve the user's API token from the keyring
	apiToken, err := keyring.Get(service, account)
	if err != nil {
		return nil, err
	}

	// Create a new Jira transport with the user's credentials
	tp := jira.BasicAuthTransport{
		Username: account, // Username is the Jira account
		Password: string(apiToken),
	}
	// Create a new Jira client with the transport
	client, err = jira.NewClient(tp.Client(), jiraURL)
	if err != nil {
		log.Printf("Error creating Jira client: %v", err)
		return nil, err
	}
	return client, nil
}

func init() {
	jiraCmd.PersistentFlags().StringVarP(&providedFlags.jiraURL, "jira-url", "j", "https://doctolib.atlassian.net", "Jira URL")
	jiraCmd.AddCommand(taskCmd)
	jiraCmd.AddCommand(jiraTokenCmd)
}
