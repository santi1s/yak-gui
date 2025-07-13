package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestStatusCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, statusCmd)
	assert.Equal(t, "status", statusCmd.Use)
	assert.Equal(t, "Get the status of one or all rollouts", statusCmd.Short)
}

func TestStatusCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := statusCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)

	// Test that watch flag exists
	watchFlag := statusCmd.Flags().Lookup("watch")
	assert.NotNil(t, watchFlag)
	assert.Equal(t, "w", watchFlag.Shorthand)
}

func TestGetRevisionInfo(t *testing.T) {
	tests := []struct {
		name     string
		rollout  *unstructured.Unstructured
		expected string
	}{
		{
			name:     "no revision",
			rollout:  &unstructured.Unstructured{},
			expected: "generation:0",
		},
		{
			name: "with revision annotation",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"rollout.argoproj.io/revision": "5",
						},
						"generation": int64(10),
					},
				},
			},
			expected: "revision:5",
		},
		{
			name: "fallback to generation",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"generation": int64(3),
					},
				},
			},
			expected: "generation:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRevisionInfo(tt.rollout)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRolloutStatus(t *testing.T) {
	// Test statusMap structure
	status := statusMap{
		Name:        "test-rollout",
		Namespace:   "default",
		Status:      "Healthy",
		Strategy:    "Canary",
		Replicas:    "3/3",
		Updated:     "3",
		Ready:       "3",
		Available:   "3",
		CurrentStep: "8/8",
		Revision:    "revision:5",
		Message:     "Rollout is healthy",
	}

	assert.Equal(t, "test-rollout", status.Name)
	assert.Equal(t, "default", status.Namespace)
	assert.Equal(t, "Healthy", status.Status)
	assert.Equal(t, "Canary", status.Strategy)
	assert.Equal(t, "3/3", status.Replicas)
	assert.Equal(t, "3", status.Updated)
	assert.Equal(t, "3", status.Ready)
	assert.Equal(t, "3", status.Available)
	assert.Equal(t, "8/8", status.CurrentStep)
	assert.Equal(t, "revision:5", status.Revision)
	assert.Equal(t, "Rollout is healthy", status.Message)
}
