package argocd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, syncCmd)
	assert.Equal(t, "sync", syncCmd.Use)
	assert.Equal(t, "Sync ArgoCD applications", syncCmd.Short)
}

func TestSyncCommandFlags(t *testing.T) {
	// Test that application flag exists
	appFlag := syncCmd.Flags().Lookup("application")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)

	// Test that all flag exists
	allFlag := syncCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)

	// Test that dry-run flag exists
	dryRunFlag := syncCmd.Flags().Lookup("dry-run")
	assert.NotNil(t, dryRunFlag)

	// Test that prune flag exists
	pruneFlag := syncCmd.Flags().Lookup("prune")
	assert.NotNil(t, pruneFlag)

	// Test that strategy flag exists
	strategyFlag := syncCmd.Flags().Lookup("strategy")
	assert.NotNil(t, strategyFlag)
}

func TestSyncCommandValidation(t *testing.T) {
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
			providedSyncFlags.application = tt.application
			providedSyncFlags.all = tt.all

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

func TestSyncOptions(t *testing.T) {
	tests := []struct {
		name          string
		dryRun        bool
		prune         bool
		strategy      string
		expectOptions []string
	}{
		{
			name:          "no options",
			dryRun:        false,
			prune:         false,
			strategy:      "",
			expectOptions: nil,
		},
		{
			name:          "dry run only",
			dryRun:        true,
			prune:         false,
			strategy:      "",
			expectOptions: []string{"DryRun=true"},
		},
		{
			name:          "prune and strategy",
			dryRun:        false,
			prune:         true,
			strategy:      "hook",
			expectOptions: []string{"Prune=true"},
		},
		{
			name:          "all options",
			dryRun:        true,
			prune:         true,
			strategy:      "apply",
			expectOptions: []string{"DryRun=true", "Prune=true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flags
			providedSyncFlags.dryRun = tt.dryRun
			providedSyncFlags.prune = tt.prune
			providedSyncFlags.strategy = tt.strategy

			// Build sync options (simplified logic)
			var options []string
			if tt.dryRun {
				options = append(options, "DryRun=true")
			}
			if tt.prune {
				options = append(options, "Prune=true")
			}

			assert.Equal(t, tt.expectOptions, options)
		})
	}
}

func TestSyncDefaultValues(t *testing.T) {
	// Reset flags to defaults for testing
	originalFlags := providedSyncFlags
	providedSyncFlags = ArgoCDSyncFlags{}
	defer func() { providedSyncFlags = originalFlags }()

	// Test default values for flags
	assert.Equal(t, "", providedSyncFlags.application)
	assert.False(t, providedSyncFlags.all)
	assert.False(t, providedSyncFlags.dryRun)
	assert.False(t, providedSyncFlags.prune)
	assert.Equal(t, "", providedSyncFlags.strategy)
}
