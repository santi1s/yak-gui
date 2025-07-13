package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestHistoryCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, historyCmd)
	assert.Equal(t, "history", historyCmd.Use)
	assert.Equal(t, "Show rollout revision history", historyCmd.Short)
}

func TestHistoryCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := historyCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)

	// Test that revision flag exists
	revisionFlag := historyCmd.Flags().Lookup("revision")
	assert.NotNil(t, revisionFlag)
}

func TestRevisionInfo(t *testing.T) {
	// Test revisionInfo structure
	rev := revisionInfo{
		Revision:   5,
		StartedAt:  "2023-01-01T10:00:00Z",
		FinishedAt: "2023-01-01T10:05:00Z",
		Duration:   "5m",
		Phase:      "Running",
		Message:    "ReplicaSet has successfully progressed.",
	}

	assert.Equal(t, int64(5), rev.Revision)
	assert.Equal(t, "2023-01-01T10:00:00Z", rev.StartedAt)
	assert.Equal(t, "2023-01-01T10:05:00Z", rev.FinishedAt)
	assert.Equal(t, "5m", rev.Duration)
	assert.Equal(t, "Running", rev.Phase)
	assert.Equal(t, "ReplicaSet has successfully progressed.", rev.Message)
}

func TestGetCurrentRevisions(t *testing.T) {
	tests := []struct {
		name     string
		rollout  *unstructured.Unstructured
		expected []revisionInfo
	}{
		{
			name:     "no revision data",
			rollout:  &unstructured.Unstructured{},
			expected: []revisionInfo{{Revision: 0}},
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
			expected: []revisionInfo{{Revision: 5}},
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
			expected: []revisionInfo{{Revision: 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCurrentRevisions(tt.rollout)
			assert.Len(t, result, len(tt.expected))
			if len(result) > 0 && len(tt.expected) > 0 {
				assert.Equal(t, tt.expected[0].Revision, result[0].Revision)
			}
		})
	}
}

func TestExtractRolloutRevisionInfo(t *testing.T) {
	tests := []struct {
		name     string
		rs       *unstructured.Unstructured
		expected revisionInfo
	}{
		{
			name:     "no annotations",
			rs:       &unstructured.Unstructured{},
			expected: revisionInfo{},
		},
		{
			name: "with rollout revision annotation",
			rs: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"rollout.argoproj.io/revision": "5",
						},
						"name": "test-rs",
					},
				},
			},
			expected: revisionInfo{
				Revision: 5,
			},
		},
		{
			name: "with deployment revision annotation",
			rs: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"deployment.kubernetes.io/revision": "3",
						},
						"name": "test-rs",
					},
				},
			},
			expected: revisionInfo{
				Revision: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRolloutRevisionInfo(tt.rs)
			assert.Equal(t, tt.expected.Revision, result.Revision)
		})
	}
}

func TestDeduplicateRevisions(t *testing.T) {
	input := []revisionInfo{
		{Revision: 1, Phase: "old"},
		{Revision: 2, Phase: "stable"},
		{Revision: 2, Phase: "current"}, // Duplicate revision, should prefer more detailed
		{Revision: 3, Phase: "current"},
	}

	result := deduplicateRevisions(input)

	// Should have 3 unique revisions
	assert.Len(t, result, 3)

	// Should prefer the more detailed info for revision 2
	foundRev2 := false
	for _, rev := range result {
		if rev.Revision == 2 {
			foundRev2 = true
			// Should prefer "current" over "stable" (more detailed info)
			assert.Equal(t, "current", rev.Phase)
		}
	}
	assert.True(t, foundRev2, "Should find revision 2")
}
