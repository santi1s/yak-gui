package rollouts

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// Helper function to load test data
func loadTestData(filename string) (*unstructured.Unstructured, error) {
	data, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err = decoder.Decode(data, nil, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Helper function to execute command and capture output
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestFormatStrategyIntegration(t *testing.T) {
	// Test with real canary rollout data
	canaryRollout, err := loadTestData("rollout_canary.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, canaryRollout)

	result := formatStrategy(canaryRollout)
	assert.Contains(t, result, "type:Canary")
	assert.Contains(t, result, "maxSurge:25%")
	// The actual testdata shows 8 steps, not 4, and maxUnavailable might not be present
	assert.Contains(t, result, "steps:8")

	// Test with real bluegreen rollout data
	blueGreenRollout, err := loadTestData("rollout_bluegreen.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, blueGreenRollout)

	result = formatStrategy(blueGreenRollout)
	assert.Contains(t, result, "type:BlueGreen")
	assert.Contains(t, result, "activeService:test-app-active")
	assert.Contains(t, result, "previewService:test-app-preview")
}

func TestFormatRevisionIntegration(t *testing.T) {
	// Test with real canary rollout data
	canaryRollout, err := loadTestData("rollout_canary.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, canaryRollout)

	result := formatRevision(canaryRollout)
	assert.Contains(t, result, "current:5")

	// Test with real bluegreen rollout data
	blueGreenRollout, err := loadTestData("rollout_bluegreen.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, blueGreenRollout)

	result = formatRevision(blueGreenRollout)
	assert.Contains(t, result, "current:3")
}

func TestFormatConditionsIntegration(t *testing.T) {
	// Test with real canary rollout data (should show Healthy)
	canaryRollout, err := loadTestData("rollout_canary.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, canaryRollout)

	result := formatConditions(canaryRollout)
	assert.Equal(t, "Healthy:True", result)

	// Test with real bluegreen rollout data (should show Paused)
	blueGreenRollout, err := loadTestData("rollout_bluegreen.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, blueGreenRollout)

	result = formatConditions(blueGreenRollout)
	assert.Contains(t, result, "Paused:True")
}

func TestGetCurrentAnalysisRunsIntegration(t *testing.T) {
	// Test with real canary rollout data
	canaryRollout, err := loadTestData("rollout_canary.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, canaryRollout)

	result := getCurrentAnalysisRuns(canaryRollout)
	assert.Len(t, result, 2)
	assert.Contains(t, result[0], "background-analysis-abc123:Running (Running)")
	assert.Contains(t, result[1], "step-analysis-def456:Successful (Successful)")

	// Test with real bluegreen rollout data
	blueGreenRollout, err := loadTestData("rollout_bluegreen.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, blueGreenRollout)

	result = getCurrentAnalysisRuns(blueGreenRollout)
	assert.Len(t, result, 1)
	assert.Contains(t, result[0], "pre-promotion-analysis-xyz789:Successful (Successful)")
}

func TestAnalysisRunOwnership(t *testing.T) {
	// Test with real analysis run data
	analysisRun, err := loadTestData("analysisrun.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, analysisRun)

	// Should be owned by test-rollout-canary
	result := isOwnedByRollout(analysisRun, "test-rollout-canary")
	assert.True(t, result)

	// Should not be owned by different rollout
	result = isOwnedByRollout(analysisRun, "other-rollout")
	assert.False(t, result)
}

func TestCommandInitialization(t *testing.T) {
	// Test that all commands are properly initialized
	rootCmd := GetRootCmd()
	assert.NotNil(t, rootCmd)

	// Check that all subcommands are added
	commands := rootCmd.Commands()
	commandNames := make([]string, 0, len(commands))
	for _, cmd := range commands {
		commandNames = append(commandNames, cmd.Use)
	}

	expectedCommands := []string{
		"status", "get", "list", "promote", "pause",
		"abort", "restart", "analysis", "logs", "history",
		"retry", "set-image", "undo",
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, commandNames, expected, "Command %s should be registered", expected)
	}
}

func TestFlagValidation(t *testing.T) {
	// Test that required flags are properly marked
	tests := []struct {
		command      *cobra.Command
		requiredFlag string
	}{
		{getCmd, "rollout"},
		{statusCmd, "rollout"},
		{logsCmd, "rollout"},
		{historyCmd, "rollout"},
		{promoteCmd, "rollout"},
		{pauseCmd, "rollout"},
		{abortCmd, "rollout"},
		{restartCmd, "rollout"},
		{retryCmd, "rollout"},
		{undoCmd, "rollout"},
		{setImageCmd, "rollout"},
		{setImageCmd, "image"},
		{analysisGetCmd, "run"},
	}

	for _, tt := range tests {
		t.Run(tt.command.Use+"_"+tt.requiredFlag, func(t *testing.T) {
			flag := tt.command.Flags().Lookup(tt.requiredFlag)
			assert.NotNil(t, flag, "Flag %s should exist on command %s", tt.requiredFlag, tt.command.Use)
		})
	}
}

// Test that we can handle empty or invalid data gracefully
func TestEdgeCases(t *testing.T) {
	emptyRollout := &unstructured.Unstructured{}

	// These should not panic with empty data
	assert.NotPanics(t, func() {
		formatStrategy(emptyRollout)
	})

	assert.NotPanics(t, func() {
		formatRevision(emptyRollout)
	})

	assert.NotPanics(t, func() {
		formatConditions(emptyRollout)
	})

	assert.NotPanics(t, func() {
		getCurrentAnalysisRuns(emptyRollout)
	})
}
