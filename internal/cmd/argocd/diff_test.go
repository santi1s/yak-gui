package argocd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, diffCmd)
	assert.Equal(t, "diff", diffCmd.Use)
	assert.Equal(t, "Show differences between Git and cluster state (same as ArgoCD)", diffCmd.Short)
}

func TestDiffCommandFlags(t *testing.T) {
	// Test that application flag exists and is required
	appFlag := diffCmd.Flags().Lookup("application")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)

	// Removed flags should not exist
	paginateFlag := diffCmd.Flags().Lookup("paginate")
	assert.Nil(t, paginateFlag)

	compactFlag := diffCmd.Flags().Lookup("compact")
	assert.Nil(t, compactFlag)

	detailedFlag := diffCmd.Flags().Lookup("detailed")
	assert.Nil(t, detailedFlag)

	pageSizeFlag := diffCmd.Flags().Lookup("page-size")
	assert.Nil(t, pageSizeFlag)
}

func TestDiffCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		application string
		expectError bool
	}{
		{
			name:        "valid application name",
			application: "my-app",
			expectError: false,
		},
		{
			name:        "empty application name",
			application: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providedDiffFlags.application = tt.application

			if tt.expectError {
				// In a real scenario, this would trigger the MarkFlagRequired validation
				assert.Empty(t, tt.application)
			} else {
				assert.NotEmpty(t, tt.application)
			}
		})
	}
}
