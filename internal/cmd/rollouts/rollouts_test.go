package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRolloutsCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, rolloutsCmd)
	assert.Equal(t, "rollouts", rolloutsCmd.Use)
	assert.Equal(t, "A suite of commands to manage Argo Rollouts resources", rolloutsCmd.Short)
}

func TestRolloutsCommandFlags(t *testing.T) {
	// Test persistent flags
	serverFlag := rolloutsCmd.PersistentFlags().Lookup("server")
	assert.NotNil(t, serverFlag)
	assert.Equal(t, "s", serverFlag.Shorthand)

	namespaceFlag := rolloutsCmd.PersistentFlags().Lookup("namespace")
	assert.NotNil(t, namespaceFlag)
	assert.Equal(t, "n", namespaceFlag.Shorthand)

	jsonFlag := rolloutsCmd.PersistentFlags().Lookup("json")
	assert.NotNil(t, jsonFlag)

	yamlFlag := rolloutsCmd.PersistentFlags().Lookup("yaml")
	assert.NotNil(t, yamlFlag)
}

func TestGetRootCmd(t *testing.T) {
	rootCmd := GetRootCmd()
	assert.NotNil(t, rootCmd)
	assert.Equal(t, rolloutsCmd, rootCmd)
}

func TestResolveNamespace(t *testing.T) {
	// Test with explicit namespace
	providedFlags.namespace = "test-namespace"
	result := resolveNamespace()
	assert.Equal(t, "test-namespace", result)

	// Test with empty namespace (should fall back to default)
	providedFlags.namespace = ""
	result = resolveNamespace()
	assert.Equal(t, "default", result)
}
