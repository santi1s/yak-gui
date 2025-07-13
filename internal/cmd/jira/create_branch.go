package jira

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	c "github.com/fatih/color"
	log "github.com/sirupsen/logrus"

	"github.com/doctolib/yak/cli"
	git "github.com/go-git/go-git/v5"

	helper "github.com/doctolib/yak/internal/helper"
	"github.com/spf13/cobra"
)

const (
	service              = "Jira API Token"
	InProgressTransition = "31"

// Transition ID: 41, Name: Done
// Transition ID: 51, Name: To be reviewed
// Transition ID: 61, Name: Idea box
// Transition ID: 71, Name: Won't do
// Transition ID: 81, Name: On hold
// Transition ID: 91, Name: To Do
)

// flags used in the create-branch command
type tasksCreateBranchFlags struct {
	TaskID     string `long:"task-id" description:"Jira task ID, if not provided a picker will be displayed to select a task from the current users in progress tasks" env:"YAK_TASK_ID"`
	ProjectKey string `long:"project-key" description:"Jira project key, if not provided it will be extracted from the current git remote" env:"YAK_JIRA_PROJECT_KEY"`
	CustomName bool   `long:"custom-name" description:"Prompt for custom branch name with Jira task ID prefix"`
}

var (
	providedBranchFlags tasksCreateBranchFlags

	createBranchCmd = &cobra.Command{
		Use:     "create-branch",
		Aliases: []string{"cb"},
		Short:   "Create branch from Jira Task",
		RunE:    createBranch,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PreRunE != nil {
				return cmd.Parent().PreRunE(cmd, args)
			}
			return nil
		},
		Example: "yak jira task cb -p PSRE -t PSRE-1111\nyak jira task cb -p PSRE",
	}
)

// Sanitize takes a string and replaces any characters that are not allowed
// in a branch name with a replacement string.
// The following characters are replaced:
//   - any character that is in the following list of characters: ['":()[\],*]
//   - any space, forward slash, backslash or backtick
//   - the characters &, ?, !, ., [, ]
//
// The function also removes any characters that are in the following list: ., !, ?
// The resulting string is returned.
func Sanitize(input string) string {
	patterns := map[*regexp.Regexp]string{
		regexp.MustCompile(`['\":()[\\],*]`):     "",  // remove these characters
		regexp.MustCompile(`[ \/\\` + "`" + `]`): "_", // replace spaces, forward slashes, backslashes and backticks with underscores
		regexp.MustCompile(`[&?!.]`):             "",  // remove these characters
	}

	for pattern, replacement := range patterns {
		input = pattern.ReplaceAllString(input, replacement)
	}

	invalidCharactersPattern := `[.!?\[\]]` // remove these characters
	input = regexp.MustCompile(invalidCharactersPattern).ReplaceAllString(input, "")

	return input
}

// promptForCustomBranchName prompts the user to enter a custom branch name
// with the provided Jira task key as a prefix
func promptForCustomBranchName(taskKey string) (string, error) {
	cli.Printf("\nComplete the branch name: %s/", taskKey)
	customName, err := cli.ReadLine()
	if err != nil {
		return "", err
	}

	// Trim whitespace and validate input
	customName = strings.TrimSpace(customName)
	if customName == "" {
		return "", fmt.Errorf("custom branch name cannot be empty")
	}

	// Sanitize the custom name
	sanitizedName := Sanitize(customName)

	return fmt.Sprintf("%s/%s", taskKey, sanitizedName), nil
}

// createBranch creates a new branch based on a Jira task
//
// The function searches for Jira tasks in the provided project
// key that the user is assigned to. If the --task-id flag is not
// provided it displays a selection menu with the tasks. If the user
// selects "None" the function exits without creating a branch.
//
// If the task is not in "In progress" state the function prompts the
// user to transition the task to "In progress" if they confirm.
//
// The function creates a new branch with the task's key and summary
// and checks it out.
func createBranch(_ *cobra.Command, _ []string) error {
	var branchName string

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

	// If the --task-id flag is not provided display a selection
	// menu with the tasks
	if providedBranchFlags.TaskID == "" {
		jql := fmt.Sprintf("project = %s AND assignee = '%s' AND issuetype = 'Task' AND status not in (Done, \"Won't do\") ORDER BY id", providedBranchFlags.ProjectKey, account)
		tasks, _, err := jiraClient.Issue.Search(jql, nil)
		if err != nil {
			log.Errorf("Error searching tasks: %v", err)
			return err
		}
		providedBranchFlags.TaskID, err = TaskSelectMenu(&tasks)
		if err != nil {
			log.Errorf("Error selecting task: %v", err)
			return err
		}
	}
	// If the user selects "None" exit without creating a branch
	if providedBranchFlags.TaskID == "0" {
		return nil
	}

	// Get the task using the provided task ID
	task, _, err := jiraClient.Issue.Get(GetIssueIDFromKey(jiraClient.Issue, providedBranchFlags.ProjectKey, providedBranchFlags.TaskID), nil)
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

	// Create a new branch with the task's key and summary or custom name
	if providedBranchFlags.CustomName {
		branchName, err = promptForCustomBranchName(task.Key)
		if err != nil {
			log.Errorf("Error getting custom branch name: %v", err)
			return err
		}
	} else {
		branchName = fmt.Sprintf("%s/%s", task.Key, Sanitize(task.Fields.Summary))
	}

	// Get the current Git repository
	dir, _ := os.Getwd()
	r, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		log.Errorf("Error opening git repository: %v", err)
		return err
	}

	// Check out the new branch
	_, err = helper.CheckoutGitBranch(r, branchName, true, true)
	if err != nil {
		log.Errorf("Error checking out branch: %v", err)
		return err
	}

	return nil
}

func init() {
	defaultProjectKey := helper.GetEnvOrDefault("YAK_JIRA_PROJECT_KEY", "PSRE")
	createBranchCmd.Flags().StringVarP(&providedBranchFlags.TaskID, "task-id", "t", "", "Jira Task ID associated with branch")
	createBranchCmd.Flags().StringVarP(&providedBranchFlags.ProjectKey, "project-key", "p", defaultProjectKey, "Jira Project key associated with branch")
	createBranchCmd.Flags().BoolVar(&providedBranchFlags.CustomName, "custom-name", false, "Prompt for custom branch name with Jira task ID prefix")
}
