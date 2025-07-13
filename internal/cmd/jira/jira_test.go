package jira

import (
	"errors"
	"testing"

	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
)

type mockGitConfig struct{}

func (g *mockGitConfig) LoadConfig(scope config.Scope) (*config.Config, error) {
	// Implement the logic to load the Git configuration here
	return nil, errors.New("Config Not found")
}

func TestGetGitConfigEmail(t *testing.T) {
	// Call the GetIssueIDFromKey function with a sample ID
	email, err := GetGitConfigEmail(&mockGitConfig{})

	// Assert the expected result
	expectedEmail := ""
	expectedErr := errors.New("Config Not found")
	assert.Equal(t, expectedEmail, email)
	assert.Equal(t, expectedErr, err)
}
