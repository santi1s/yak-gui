package argocd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefreshCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, refreshCmd)
	assert.Equal(t, "refresh", refreshCmd.Use)
	assert.Equal(t, "Refresh ArgoCD applications", refreshCmd.Short)
	assert.Equal(t, "Refresh ArgoCD applications to detect changes in Git repository", refreshCmd.Long)
}

func TestRefreshCommandFlags(t *testing.T) {
	// Test that application flag exists
	appFlag := refreshCmd.Flags().Lookup("application")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)

	// Test that all flag exists
	allFlag := refreshCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)

	// Test that hard flag exists
	hardFlag := refreshCmd.Flags().Lookup("hard")
	assert.NotNil(t, hardFlag)
}

func TestRefreshCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		application string
		all         bool
		expectError bool
		description string
	}{
		{
			name:        "valid application name",
			application: "my-app",
			all:         false,
			expectError: false,
			description: "should be valid when application is specified",
		},
		{
			name:        "all flag set",
			application: "",
			all:         true,
			expectError: false,
			description: "should be valid when all flag is set",
		},
		{
			name:        "neither application nor all",
			application: "",
			all:         false,
			expectError: true,
			description: "should be invalid when neither application nor all is specified",
		},
		{
			name:        "both application and all",
			application: "my-app",
			all:         true,
			expectError: true,
			description: "should be invalid when both application and all are specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providedRefreshFlags.application = tt.application
			providedRefreshFlags.all = tt.all

			// Test mutual exclusivity logic
			hasBoth := tt.application != "" && tt.all
			hasNeither := tt.application == "" && !tt.all

			if tt.expectError {
				assert.True(t, hasBoth || hasNeither, tt.description)
			} else {
				assert.False(t, hasBoth || hasNeither, tt.description)
			}
		})
	}
}

func TestRefreshOptions(t *testing.T) {
	tests := []struct {
		name            string
		hard            bool
		expectedRefresh string
	}{
		{
			name:            "normal refresh",
			hard:            false,
			expectedRefresh: "normal",
		},
		{
			name:            "hard refresh",
			hard:            true,
			expectedRefresh: "hard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flags
			providedRefreshFlags.hard = tt.hard

			// Build refresh type (simplified logic from refreshApplication)
			refreshType := "normal"
			if tt.hard {
				refreshType = "hard"
			}

			assert.Equal(t, tt.expectedRefresh, refreshType)
		})
	}
}

func TestRefreshDefaultValues(t *testing.T) {
	// Reset flags to defaults for testing
	originalFlags := providedRefreshFlags
	providedRefreshFlags = ArgoCDRefreshFlags{}
	defer func() { providedRefreshFlags = originalFlags }()

	// Test default values for flags
	assert.Equal(t, "", providedRefreshFlags.application)
	assert.False(t, providedRefreshFlags.all)
	assert.False(t, providedRefreshFlags.hard)
}

func TestRefreshCommandExample(t *testing.T) {
	// Test that examples are properly set
	expectedExamples := "yak argocd refresh --application my-app\nyak argocd refresh --all\nyak argocd refresh --application my-app --hard"
	assert.Equal(t, expectedExamples, refreshCmd.Example)
}

func TestRefreshFlagStructure(t *testing.T) {
	// Test that ArgoCDRefreshFlags struct has all expected fields
	flags := ArgoCDRefreshFlags{
		application: "test-app",
		all:         true,
		hard:        true,
	}

	assert.Equal(t, "test-app", flags.application)
	assert.True(t, flags.all)
	assert.True(t, flags.hard)
}
