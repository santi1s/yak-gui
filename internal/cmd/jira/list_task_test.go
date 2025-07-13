package jira

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommaSeparatedStatusList(t *testing.T) {
	// Invoke the commaSeparatedStatusList function
	result := commaSeparatedStatusList([]string{"Open", "In Progress", "Done"})

	// Assert the expected result
	expected := "\"Open\", \"In Progress\", \"Done\""
	assert.Equal(t, expected, result)
}

func TestCommaSeparatedStatusList_EmptyList(t *testing.T) {
	// Invoke the commaSeparatedStatusList function with an empty list
	result := commaSeparatedStatusList([]string{})

	// Assert the expected result
	expected := ""
	assert.Equal(t, expected, result)
}
