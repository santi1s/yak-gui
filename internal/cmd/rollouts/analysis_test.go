package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestAnalysisCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, analysisCmd)
	assert.Equal(t, "analysis", analysisCmd.Use)
	assert.Equal(t, "Manage analysis templates and runs", analysisCmd.Short)
}

func TestAnalysisGetCommand(t *testing.T) {
	// Test get subcommand structure
	assert.NotNil(t, analysisGetCmd)
	assert.Equal(t, "get", analysisGetCmd.Use)
	assert.Equal(t, "Get analysis template or run details", analysisGetCmd.Short)
}

func TestAnalysisGetCommandFlags(t *testing.T) {
	// Test that run flag exists and is required
	runFlag := analysisGetCmd.Flags().Lookup("run")
	assert.NotNil(t, runFlag)
	assert.Equal(t, "r", runFlag.Shorthand)
}

func TestAnalysisListCommand(t *testing.T) {
	// Test list subcommand structure
	assert.NotNil(t, analysisListCmd)
	assert.Equal(t, "list", analysisListCmd.Use)
	assert.Equal(t, "List analysis templates or runs", analysisListCmd.Short)
}

func TestAnalysisListCommandFlags(t *testing.T) {
	// Test that all flag exists
	allFlag := analysisListCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
}

func TestFormatAnalysisRunDetails(t *testing.T) {
	// Test that the function exists and can handle basic cases
	run := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test-analysis-run",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"phase": "Successful",
			},
		},
	}

	// This function prints output, so we can't easily test the output
	// but we can ensure it doesn't panic
	assert.NotPanics(t, func() {
		formatAnalysisRunDetails(run)
	})
}

func TestAnalysisRunFlags(t *testing.T) {
	// Test RolloutsAnalysisFlags structure
	flags := RolloutsAnalysisFlags{
		run: "test-run",
	}
	assert.Equal(t, "test-run", flags.run)

	// Test RolloutsAnalysisFlags structure for list
	listFlags := RolloutsAnalysisFlags{
		all:      true,
		template: "test-template",
	}
	assert.Equal(t, true, listFlags.all)
	assert.Equal(t, "test-template", listFlags.template)
}
