package jira

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	jira "github.com/andygrunwald/go-jira"
	log "github.com/sirupsen/logrus"
)

type JiraIssueInterface interface {
	Search(jql string, opts *jira.SearchOptions) ([]jira.Issue, *jira.Response, error)
}

// GetIssueIDFromKey retrieves the ID of the issue with the given key
// from Jira.
//
// The JQL query is constructed using the project key provided
// through the command line flags.
//
// The function logs an error if there is an error searching for the
// issue or if no issue is found.
//
// The function returns an empty string if there is an error or no
// issue is found.
func GetIssueIDFromKey(j JiraIssueInterface, projectKey, id string) string {
	jql := fmt.Sprintf("project = %s AND key = '%s'", projectKey, id)
	issues, _, err := j.Search(jql, nil)
	if err != nil || len(issues) == 0 {
		log.Errorf("Error searching tasks: %v with project name: %v, ID: %v", err, projectKey, id)
		return ""
	}
	return issues[0].ID
}

// TaskSelectMenu displays a menu with the given tasks and returns the
// user's selection.
//
// The menu displays the task's key, summary and status. If the user
// selects "None" the function returns "0".
func TaskSelectMenu(tasks *[]jira.Issue) (string, error) {
	var TasksSlice []string
	SelectedTask := ""
	for _, task := range *tasks {
		TasksSlice = append(TasksSlice, fmt.Sprintf("%s %s [%s]", task.Key, task.Fields.Summary, task.Fields.Status.Name))
	}

	prompt := &survey.Select{
		Message: "Select a task:",
		Options: TasksSlice,
	}
	err := survey.AskOne(prompt, &SelectedTask)
	return strings.Split(SelectedTask, " ")[0], err
}

// ExtractJiraIDFromBranchName extracts a Jira ID from the given branch name.
// It looks for a pattern of 2-10 uppercase characters followed by a dash and numbers.
// Returns the Jira ID if found, empty string otherwise.
func ExtractJiraIDFromBranchName(branchName string) string {
	// Pattern: word boundary, 2-10 uppercase letters, dash, one or more digits, word boundary
	pattern := `\b([A-Z]{2,10}-\d+)\b`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(branchName)
	if len(matches) > 1 {
		// Additional validation: ensure the project key is exactly 2-10 characters
		parts := strings.Split(matches[1], "-")
		if len(parts) >= 2 && len(parts[0]) >= 2 && len(parts[0]) <= 10 {
			return matches[1]
		}
	}

	return ""
}
