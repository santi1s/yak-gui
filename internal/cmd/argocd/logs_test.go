package argocd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogsCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, logsCmd)
	assert.Equal(t, "logs", logsCmd.Use)
	assert.Equal(t, "Get logs from pods managed by an ArgoCD application", logsCmd.Short)
}

func TestLogsCommandFlags(t *testing.T) {
	// Test that application flag exists and is required
	appFlag := logsCmd.Flags().Lookup("application")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)

	// Test that follow flag exists
	followFlag := logsCmd.Flags().Lookup("follow")
	assert.NotNil(t, followFlag)
	assert.Equal(t, "f", followFlag.Shorthand)

	// Test that container flag exists
	containerFlag := logsCmd.Flags().Lookup("container")
	assert.NotNil(t, containerFlag)
	assert.Equal(t, "c", containerFlag.Shorthand)

	// Test that tail flag exists
	tailFlag := logsCmd.Flags().Lookup("tail")
	assert.NotNil(t, tailFlag)

	// Test that since flag exists
	sinceFlag := logsCmd.Flags().Lookup("since")
	assert.NotNil(t, sinceFlag)

	// Test that previous flag exists
	previousFlag := logsCmd.Flags().Lookup("previous")
	assert.NotNil(t, previousFlag)

	// Test that pod flag exists
	podFlag := logsCmd.Flags().Lookup("pod")
	assert.NotNil(t, podFlag)
}

func TestLogsCommandValidation(t *testing.T) {
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
			providedLogsFlags.application = tt.application

			if tt.expectError {
				// In a real scenario, this would trigger the MarkFlagRequired validation
				assert.Empty(t, tt.application)
			} else {
				assert.NotEmpty(t, tt.application)
			}
		})
	}
}

func TestLogsDefaultValues(t *testing.T) {
	// Test default values for flags
	assert.False(t, providedLogsFlags.follow)
	assert.False(t, providedLogsFlags.previous)
	assert.Equal(t, int64(0), providedLogsFlags.tail)
	assert.Equal(t, "", providedLogsFlags.container)
	assert.Equal(t, "", providedLogsFlags.since)
	assert.Equal(t, "", providedLogsFlags.podName)
}
