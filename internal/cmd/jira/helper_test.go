package jira

import (
	"testing"

	jira "github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
)

type mockJiraIssueService struct{}

func (c *mockJiraIssueService) Search(jql string, opts *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
	// Mock the behavior of the Issue.Search method
	mockIssues := []jira.Issue{
		{ID: "123", Key: "TASK-1"},
		{ID: "456", Key: "TASK-2"},
	}
	return mockIssues, nil, nil
}

func TestGetIssueIDFromKey(t *testing.T) {
	// Call the GetIssueIDFromKey function with a sample ID
	issueID := GetIssueIDFromKey(&mockJiraIssueService{}, "SREGREEN", "TASK-123")

	// Assert the expected result
	expectedIssueID := "123" // Replace with the expected issue ID
	assert.Equal(t, expectedIssueID, issueID)
}

func TestExtractJiraIDFromBranchName(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		expectedID  string
		description string
	}{
		{
			name:        "valid_jira_id_with_prefix",
			branchName:  "SREGREEN-675/create-commit-from-branch-id",
			expectedID:  "SREGREEN-675",
			description: "Should extract Jira ID from branch name with prefix",
		},
		{
			name:        "valid_jira_id_start_of_branch",
			branchName:  "PSRE-123-feature-branch",
			expectedID:  "PSRE-123",
			description: "Should extract Jira ID from start of branch name",
		},
		{
			name:        "valid_jira_id_in_middle",
			branchName:  "feature/TASK-456/implementation",
			expectedID:  "TASK-456",
			description: "Should extract Jira ID from middle of branch name",
		},
		{
			name:        "short_project_key",
			branchName:  "AB-789-fix",
			expectedID:  "AB-789",
			description: "Should extract Jira ID with 2-character project key",
		},
		{
			name:        "long_project_key",
			branchName:  "MAXLENGKEY-100-test",
			expectedID:  "MAXLENGKEY-100",
			description: "Should extract Jira ID with 10-character project key",
		},
		{
			name:        "multiple_jira_ids_first_match",
			branchName:  "PROJ-123-related-to-TASK-456",
			expectedID:  "PROJ-123",
			description: "Should extract first Jira ID when multiple present",
		},
		{
			name:        "no_jira_id_lowercase",
			branchName:  "feature/test-123-branch",
			expectedID:  "",
			description: "Should return empty string for lowercase project key",
		},
		{
			name:        "no_jira_id_no_dash",
			branchName:  "TASK123-feature",
			expectedID:  "",
			description: "Should return empty string without dash separator",
		},
		{
			name:        "no_jira_id_no_number",
			branchName:  "TASK--feature",
			expectedID:  "",
			description: "Should return empty string without ticket number",
		},
		{
			name:        "no_jira_id_single_char",
			branchName:  "A-123-feature",
			expectedID:  "",
			description: "Should return empty string for single character project key",
		},
		{
			name:        "no_jira_id_too_long",
			branchName:  "VERYLONGPROJECTKEY-123-feature",
			expectedID:  "",
			description: "Should return empty string for project key longer than 10 characters",
		},
		{
			name:        "no_jira_id_plain_branch",
			branchName:  "main",
			expectedID:  "",
			description: "Should return empty string for plain branch names",
		},
		{
			name:        "no_jira_id_feature_branch",
			branchName:  "feature/new-functionality",
			expectedID:  "",
			description: "Should return empty string for feature branches without Jira ID",
		},
		{
			name:        "jira_id_with_leading_numbers",
			branchName:  "123-TASK-456-branch",
			expectedID:  "TASK-456",
			description: "Should extract Jira ID even with leading numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractJiraIDFromBranchName(tt.branchName)
			assert.Equal(t, tt.expectedID, result, tt.description)
		})
	}
}
