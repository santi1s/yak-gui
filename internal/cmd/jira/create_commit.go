package jira

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	c "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"

	"github.com/doctolib/yak/cli"
	helper "github.com/doctolib/yak/internal/helper"

	"github.com/spf13/cobra"
)

// flags used in the create-commit command
type tasksCreateCommitFlags struct {
	TaskID           string `long:"task-id" description:"Jira task ID, if not provided a picker will be displayed to select a task from the current users in progress tasks" env:"YAK_TASK_ID"`
	ProjectKey       string `long:"project-key" description:"Jira project key, if not provided it will be extracted from the current git remote" env:"YAK_JIRA_PROJECT_KEY"`
	IDFromBranchName bool   `long:"id-from-branch-name" description:"Extract Jira task ID from the current branch name instead of showing task picker"`
}

var (
	providedCommitFlags tasksCreateCommitFlags

	createCommitCmd = &cobra.Command{
		Use:     "create-commit",
		Aliases: []string{"cc"},
		Short:   "Create commit from Jira Task",
		RunE:    createCommit,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PreRunE != nil {
				return cmd.Parent().PreRunE(cmd, args)
			}
			return nil
		},
		Example: "yak jira task cc -p PSRE -t PSRE-1111\nyak jira task cb -p PSRE",
	}
)

// createCommit creates a new commit based on a Jira task, commit type,
// and description
//
// The function searches for Jira tasks in the provided project
// key that the user is assigned to. If the --task-id flag is not
// provided it displays a selection menu with the tasks. If the user
// selects "None" the function exits without creating a commit.
//
// If the task is not in "In progress" state the function prompts the
// user to transition the task to "In progress" if they confirm.
//
// If a task was successfully found, the function prompts the user to
// select a commit type and enter a commit message.
//
// The function then creates a new commit with the constructed message
func createCommit(_ *cobra.Command, _ []string) error {
	var commitType string

	// Get the current Git repository
	dir, _ := os.Getwd()
	r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		log.Errorf("Error opening git repository: %v", err)
		return err
	}

	// If no staged changes exist, it doesn't make sense to create a commit
	hasStagedChanges, err := helper.HasStagedGitChanges(r)
	if err != nil {
		log.Errorf("Error checking staged changes: %v", err)
		return err
	}
	if !hasStagedChanges {
		log.Warnf("No staged changes found. Please stage your changes before attempting to creating a commit.")
		return nil
	}

	// Get the user's email from the global Git config
	account, err := GetGitConfigEmail(&GitConfig{})
	if err != nil {
		log.Errorf("Error getting user email from the global Git config: %v", err)
		return err
	}

	// Get a Jira client
	jiraClient, err := GetJiraClient(account, providedFlags.jiraURL)
	if err != nil {
		log.Errorf("Error getting Jira client: %v", err)
		return err
	}

	// If the --task-id flag is not provided, either extract from branch name
	// (if --id-from-branch-name flag is set) or display a selection menu
	if providedCommitFlags.TaskID == "" {
		if providedCommitFlags.IDFromBranchName {
			// Get current branch name
			head, err := r.Head()
			if err != nil {
				log.Errorf("Error getting current branch: %v", err)
				return err
			}
			branchName := head.Name().Short()

			// Extract Jira ID from branch name
			extractedID := ExtractJiraIDFromBranchName(branchName)
			if extractedID != "" {
				providedCommitFlags.TaskID = extractedID
			} else {
				log.Warnf("No Jira ID found in branch name '%s', falling back to task selection menu", branchName)
			}
		}

		// If still no task ID (either flag not set or no ID found in branch name),
		// display selection menu
		if providedCommitFlags.TaskID == "" {
			jql := fmt.Sprintf("project = %s AND assignee = '%s' AND issuetype = 'Task' AND status not in (Done, \"Won't do\") ORDER BY id", providedCommitFlags.ProjectKey, account)
			tasks, _, err := jiraClient.Issue.Search(jql, nil)
			if err != nil {
				log.Errorf("Error searching tasks: %v", err)
				return err
			}
			providedCommitFlags.TaskID, err = TaskSelectMenu(&tasks)
			if err != nil {
				log.Errorf("Error selecting task: %v", err)
				return err
			}
		}
	}
	// If the user selects "None" exit without creating a commit
	if providedCommitFlags.TaskID == "0" {
		return nil
	}

	// Get the task using the provided task ID
	task, _, err := jiraClient.Issue.Get(GetIssueIDFromKey(jiraClient.Issue, providedCommitFlags.ProjectKey, providedCommitFlags.TaskID), nil)
	if err != nil {
		log.Errorf("Error searching task: %v", err)
		return err
	}

	// If the task is not in "In progress" state prompt the user to
	// transition the task to "In progress" if they confirm
	if strings.ToLower(task.Fields.Status.Name) != "in progress" {
		confirmationPrompt := fmt.Sprintf("Task is in state %s. Transition task to In progress?", c.New(c.FgYellow).SprintFunc()(task.Fields.Status.Name))
		if cli.AskConfirmation(confirmationPrompt) {
			_, err = jiraClient.Issue.DoTransition(task.ID, InProgressTransition)
			if err != nil {
				log.Errorf("Error transioning task: %v", err)
				return err
			}
		}
	}

	// Prompt the user to select a commit type
	commitTypePrompt := &survey.Select{
		Message: "Choose a commit type:",
		Options: []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"},
	}

	err = survey.AskOne(commitTypePrompt, &commitType)
	if err != nil {
		log.Errorf("Error selecting commit type: %v", err)
		return err
	}

	// Prompt the user to enter a commit description
	commitDescPrompt := &survey.Input{
		Message: fmt.Sprintf("Complete the commit message:\n\t%s(%s):", commitType, task.Key),
	}

	var commitDesc string
	err = survey.AskOne(commitDescPrompt, &commitDesc)
	if err != nil {
		log.Errorf("Error entering commit message: %v", err)
		return err
	}

	// Construct the commit message
	commitMsg := fmt.Sprintf("%s(%s): %s\n", commitType, task.Key, strings.TrimSpace(commitDesc)) // TrimSpace to remove leading/trailing spaces

	// Create the commit
	err = helper.CreateGitCommit(r, commitMsg)
	if err != nil {
		log.Errorf("Error creating commit: %v", err)
		return err
	}

	fmt.Print("\nDone!\n\nDid you commit something by mistake and safely want to revert what you just did? Run `git reset HEAD~`.\n")

	return nil
}

func init() {
	defaultProjectKey := helper.GetEnvOrDefault("YAK_JIRA_PROJECT_KEY", "PSRE")
	createCommitCmd.Flags().StringVarP(&providedCommitFlags.TaskID, "task-id", "t", "", "Jira Task ID associated with branch")
	createCommitCmd.Flags().StringVarP(&providedCommitFlags.ProjectKey, "project-key", "p", defaultProjectKey, "Jira Project key associated with branch")
	createCommitCmd.Flags().BoolVar(&providedCommitFlags.IDFromBranchName, "id-from-branch-name", false, "Extract Jira task ID from the current branch name instead of showing task picker")
}
