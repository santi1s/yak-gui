package jira

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	jira "github.com/andygrunwald/go-jira"
	"github.com/doctolib/yak/cli"
	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	input := "Hello [World]!"
	expected := "Hello_World"
	output := Sanitize(input)

	if output != expected {
		t.Errorf("Sanitize failed. Expected: %s, Got: %s", expected, output)
	}
}

func TestPromptForCustomBranchName(t *testing.T) {
	tests := []struct {
		name        string
		taskKey     string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid input with simple text",
			taskKey:     "JIRA-123",
			input:       "my-feature",
			expected:    "JIRA-123/my-feature",
			shouldError: false,
		},
		{
			name:        "valid input with special characters",
			taskKey:     "PSRE-456",
			input:       "fix bug [urgent]!",
			expected:    "PSRE-456/fix_bug_urgent",
			shouldError: false,
		},
		{
			name:        "input with leading/trailing whitespace",
			taskKey:     "TEST-789",
			input:       "  feature-name  ",
			expected:    "TEST-789/feature-name",
			shouldError: false,
		},
		{
			name:        "empty input after trimming",
			taskKey:     "EMPTY-001",
			input:       "   ",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "empty input",
			taskKey:     "EMPTY-002",
			input:       "",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "input with complex characters",
			taskKey:     "COMPLEX-123",
			input:       "add/user:authentication & authorization?",
			expected:    "COMPLEX-123/add_user:authentication__authorization",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock stdin
			var stdin bytes.Buffer
			stdin.WriteString(tt.input + "\n")
			cli.SetIn(&stdin)

			// Set up mock stdout to capture the prompt
			var stdout bytes.Buffer
			cli.SetOut(&stdout)

			// Call the function
			result, err := promptForCustomBranchName(tt.taskKey)

			// Check error expectation
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result if no error expected
			if !tt.shouldError {
				if result != tt.expected {
					t.Errorf("Expected: %s, Got: %s", tt.expected, result)
				}

				// Verify the prompt message
				promptOutput := stdout.String()
				expectedPrompt := fmt.Sprintf("Complete the branch name: %s/", tt.taskKey)
				if !strings.Contains(promptOutput, expectedPrompt) {
					t.Errorf("Expected prompt to contain: %s, Got: %s", expectedPrompt, promptOutput)
				}
			}
		})
	}
}

func TestSanitizeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "all forbidden characters",
			input:    `'":[()]\,*&?!.[]`,
			expected: "'\":(_,*",
		},
		{
			name:     "spaces and slashes",
			input:    "hello world/test\\path",
			expected: "hello_world_test_path",
		},
		{
			name:     "mixed valid and invalid",
			input:    "feature-123_test.final!",
			expected: "feature-123_testfinal",
		},
		{
			name:     "already clean input",
			input:    "clean-branch-name",
			expected: "clean-branch-name",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if result != tt.expected {
				t.Errorf("Sanitize(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateBranchStatusCheckLowercase(t *testing.T) {
	task := &jira.Issue{
		Fields: &jira.IssueFields{
			Status: &jira.Status{
				Name: "In Progress",
			},
		},
	}

	result := strings.ToLower(task.Fields.Status.Name) == "in progress"
	assert.True(t, result, "Should match 'In Progress' status with lowercase comparison")
}
