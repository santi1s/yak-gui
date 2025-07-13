package jira

import (
	"strings"
	"testing"

	jira "github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
)

func TestStatusCheckLowercase(t *testing.T) {
	task := &jira.Issue{
		Fields: &jira.IssueFields{
			Status: &jira.Status{
				Name: "IN PROGRESS",
			},
		},
	}

	result := strings.ToLower(task.Fields.Status.Name) == "in progress"
	assert.True(t, result, "Should match 'In Progress' status with lowercase comparison")
}
