package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	assert.Equal(t, "List rollouts", listCmd.Short)
}

func TestListCommandFlags(t *testing.T) {
	// Test that all flag exists
	allFlag := listCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
}

func TestListFlags(t *testing.T) {
	// Test default values
	flags := RolloutsListFlags{}
	assert.Equal(t, false, flags.all)
}
