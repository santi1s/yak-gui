package jira

import (
	"fmt"
	"os"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/doctolib/yak/internal/helper"
	log "github.com/sirupsen/logrus"

	"github.com/AlecAivazis/survey/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type tasksListFlags struct {
	projectKey string
	status     bool
	all        bool
}

var (
	providedListFlags tasksListFlags

	listTasksCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List Jira Tasks",
		RunE:    listTasks,
		Example: "yak jira l -p PSRE -s\nyak jira l -p PSRE -a",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().PreRunE != nil {
				return cmd.Parent().PreRunE(cmd, args)
			}
			return nil
		},
	}

	TaskStatus = []string{"New", "Ready", "In progress", "On hold", "To be reviewed", "To Do"}
)

// printTasks prints the given data in a table format to stdout
func printTasks(data [][]string) {
	// Create a new table with some optional features enabled
	table := tablewriter.NewWriter(os.Stdout)

	// Set the headers of the table
	table.SetHeader([]string{"Task", "Descripton", "Status", "Assignee"})

	// Disable automatic wrapping of the table cells
	table.SetAutoWrapText(false)

	// Set the headers to be aligned to the left
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	// Set the table data to be aligned to the left
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	// Remove the borders from the table
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")

	// Add a line between the header and the rest of the table
	table.SetHeaderLine(true)

	// Pad the table cells with tabs
	table.SetTablePadding("\t")

	// Remove any whitespace between cells
	table.SetNoWhiteSpace(true)

	// Add the data to the table
	table.AppendBulk(data)

	// Render the table
	table.Render()
}

// commaSeparatedStatusList formats a slice of strings into a comma-separated
// list of quoted strings.
//
// The returned string is safe to be used directly in a JQL query.
func commaSeparatedStatusList(statusSlice []string) string {
	formattedSlice := make([]string, len(statusSlice))
	for i, str := range statusSlice {
		formattedSlice[i] = fmt.Sprintf(`"%s"`, str)
	}
	return strings.Join(formattedSlice, ", ")
}

// listTasks lists all the tasks in a given Jira project
func listTasks(cmd *cobra.Command, _ []string) error {
	// Get the Git config email as the user
	account, err := GetGitConfigEmail(&GitConfig{})

	if err != nil {
		log.Errorf("Error getting user email from the global Git config: %v", err)
		return err
	}

	// Get the Jira client
	jiraClient, err := GetJiraClient(account, providedFlags.jiraURL)
	if err != nil {
		log.Errorf("Error getting Jira client: %v", err)
		return err
	}

	// Build the JQL query
	var jql string
	if cmd.Flags().Changed("status") && providedListFlags.status {
		// Get the user's input for the task statuses
		prompt := &survey.MultiSelect{
			Message: "Select status:",
			Options: TaskStatus,
		}
		var selectedStatuses []string
		err := survey.AskOne(prompt, &selectedStatuses)
		if err != nil {
			log.Errorf("Failed to prompt: %v", err)
			return err
		}
		if len(selectedStatuses) == 0 {
			fmt.Println("No status selected")
			return nil
		}
		jql = `issuetype = 'Task' AND status in (` + commaSeparatedStatusList(selectedStatuses) + `)`
		jql = jql + fmt.Sprintf(" AND project = %s", providedListFlags.projectKey)
	} else {
		jql = `issuetype = 'Task' AND status not in (Done, "Won't do")`
		jql = jql + fmt.Sprintf(" AND project = %s", providedListFlags.projectKey)
	}
	if !providedListFlags.all {
		jql = jql + fmt.Sprintf(" AND assignee = '%s'", account)
	}

	// Search for the issues
	options := jira.SearchOptions{
		MaxResults: 200,
	}
	issues, _, err := jiraClient.Issue.Search(jql, &options)
	if err != nil {
		log.Errorf("Failed to search for issues: %v", err)
		return err
	}

	// Create a slice of data to be printed
	var taskData [][]string

	// Print the list of issues
	for _, issue := range issues {
		if issue.Fields.Assignee != nil {
			taskData = append(taskData, []string{issue.Key, issue.Fields.Summary, issue.Fields.Status.Name, issue.Fields.Assignee.DisplayName})
		} else {
			taskData = append(taskData, []string{issue.Key, issue.Fields.Summary, issue.Fields.Status.Name, "Unassigned"})
		}
	}

	if len(taskData) > 0 {
		printTasks(taskData)
	} else {
		fmt.Println("No tasks found")
	}
	return nil
}

func init() {
	defaultProjectKey := helper.GetEnvOrDefault("YAK_JIRA_PROJECT_KEY", "PSRE")
	listTasksCmd.Flags().StringVarP(&providedListFlags.projectKey, "project-key", "p", defaultProjectKey, "Jira Project key")
	listTasksCmd.Flags().BoolVarP(&providedListFlags.status, "status", "s", false, "Use status filter")
	listTasksCmd.Flags().BoolVarP(&providedListFlags.all, "all", "a", false, "Show tasks from all assignees and not only the logged user")
}
