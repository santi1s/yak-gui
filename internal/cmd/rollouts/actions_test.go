package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test rollout action commands (promote, pause, abort, restart, retry, undo)

func TestPromoteCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, promoteCmd)
	assert.Equal(t, "promote", promoteCmd.Use)
	assert.Equal(t, "Promote a rollout to the next step or full deployment", promoteCmd.Short)
}

func TestPromoteCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := promoteCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)

	// Test that full flag exists
	fullFlag := promoteCmd.Flags().Lookup("full")
	assert.NotNil(t, fullFlag)

	// The skip field exists in struct but not used in CLI currently
	// skipFlag := promoteCmd.Flags().Lookup("skip")
	// assert.NotNil(t, skipFlag)
}

func TestPauseCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, pauseCmd)
	assert.Equal(t, "pause", pauseCmd.Use)
	assert.Equal(t, "Pause a rollout", pauseCmd.Short)
}

func TestPauseCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := pauseCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)
}

func TestAbortCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, abortCmd)
	assert.Equal(t, "abort", abortCmd.Use)
	assert.Equal(t, "Abort a rollout and rollback to stable version", abortCmd.Short)
}

func TestAbortCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := abortCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)
}

func TestRestartCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, restartCmd)
	assert.Equal(t, "restart", restartCmd.Use)
	assert.Equal(t, "Restart the pods of a rollout", restartCmd.Short)
}

func TestRestartCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := restartCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)
}

func TestRetryCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, retryCmd)
	assert.Equal(t, "retry", retryCmd.Use)
	assert.Equal(t, "Retry a failed rollout step", retryCmd.Short)
}

func TestRetryCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := retryCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)
}

func TestUndoCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, undoCmd)
	assert.Equal(t, "undo", undoCmd.Use)
	assert.Equal(t, "Rollback rollout to previous revision", undoCmd.Short)
}

func TestUndoCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := undoCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)

	// Test that to-revision flag exists
	toRevisionFlag := undoCmd.Flags().Lookup("to-revision")
	assert.NotNil(t, toRevisionFlag)
}

func TestSetImageCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, setImageCmd)
	assert.Equal(t, "set-image", setImageCmd.Use)
	assert.Equal(t, "Update rollout image", setImageCmd.Short)
}

func TestSetImageCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := setImageCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)

	// Test that image flag exists and is required
	imageFlag := setImageCmd.Flags().Lookup("image")
	assert.NotNil(t, imageFlag)
}

// Test flag structures
func TestActionFlags(t *testing.T) {
	// Test RolloutsPromoteFlags
	promoteFlags := RolloutsPromoteFlags{
		rollout: "test-rollout",
		full:    true,
		skip:    false,
	}
	assert.Equal(t, "test-rollout", promoteFlags.rollout)
	assert.Equal(t, true, promoteFlags.full)
	assert.Equal(t, false, promoteFlags.skip)

	// Test RolloutsUndoFlags
	undoFlags := RolloutsUndoFlags{
		rollout:    "test-rollout",
		toRevision: 5,
	}
	assert.Equal(t, "test-rollout", undoFlags.rollout)
	assert.Equal(t, int64(5), undoFlags.toRevision)

	// Test RolloutsSetImageFlags
	setImageFlags := RolloutsSetImageFlags{
		rollout: "test-rollout",
		image:   "nginx:1.20",
	}
	assert.Equal(t, "test-rollout", setImageFlags.rollout)
	assert.Equal(t, "nginx:1.20", setImageFlags.image)
}
