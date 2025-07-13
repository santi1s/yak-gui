package rollouts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, getCmd)
	assert.Equal(t, "get", getCmd.Use)
	assert.Equal(t, "Get detailed information about a rollout", getCmd.Short)
}

func TestGetCommandFlags(t *testing.T) {
	// Test that rollout flag exists and is required
	rolloutFlag := getCmd.Flags().Lookup("rollout")
	assert.NotNil(t, rolloutFlag)
	assert.Equal(t, "r", rolloutFlag.Shorthand)
}

func TestFormatStrategy(t *testing.T) {
	tests := []struct {
		name     string
		rollout  *unstructured.Unstructured
		expected string
	}{
		{
			name:     "no strategy",
			rollout:  &unstructured.Unstructured{},
			expected: "<none>",
		},
		{
			name: "canary strategy",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"strategy": map[string]interface{}{
							"canary": map[string]interface{}{
								"maxSurge":       "25%",
								"maxUnavailable": "1",
								"steps": []interface{}{
									map[string]interface{}{"setWeight": "20"},
									map[string]interface{}{"pause": map[string]interface{}{}},
								},
							},
						},
					},
				},
			},
			expected: "type:Canary\nmaxSurge:25%\nmaxUnavailable:1\nsteps:2",
		},
		{
			name: "bluegreen strategy",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"spec": map[string]interface{}{
						"strategy": map[string]interface{}{
							"blueGreen": map[string]interface{}{
								"activeService":  "my-app-active",
								"previewService": "my-app-preview",
							},
						},
					},
				},
			},
			expected: "type:BlueGreen\nactiveService:my-app-active\npreviewService:my-app-preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStrategy(tt.rollout)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatRevision(t *testing.T) {
	tests := []struct {
		name     string
		rollout  *unstructured.Unstructured
		expected string
	}{
		{
			name:     "no revision",
			rollout:  &unstructured.Unstructured{},
			expected: "<none>",
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
			expected: "current:5",
		},
		{
			name: "only generation",
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
			result := formatRevision(tt.rollout)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatConditions(t *testing.T) {
	tests := []struct {
		name     string
		rollout  *unstructured.Unstructured
		expected string
	}{
		{
			name:     "no conditions",
			rollout:  &unstructured.Unstructured{},
			expected: "<none>",
		},
		{
			name: "healthy condition",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Healthy",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Healthy:True",
		},
		{
			name: "multiple conditions with filtering",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Healthy",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Progressing",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "Available",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Healthy:True", // Progressing and Available should be filtered out when Healthy is True
		},
		{
			name: "paused condition",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Paused",
								"status": "True",
							},
						},
					},
				},
			},
			expected: "Paused:True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatConditions(tt.rollout)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentAnalysisRuns(t *testing.T) {
	tests := []struct {
		name     string
		rollout  *unstructured.Unstructured
		expected []string
	}{
		{
			name:     "no analysis runs",
			rollout:  &unstructured.Unstructured{},
			expected: []string{},
		},
		{
			name: "canary background analysis run",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"canary": map[string]interface{}{
							"currentBackgroundAnalysisRuns": []interface{}{
								map[string]interface{}{
									"name":   "background-analysis-123",
									"status": "Running",
									"phase":  "Running",
								},
							},
						},
					},
				},
			},
			expected: []string{"background-analysis-123:Running (Running)"},
		},
		{
			name: "step analysis run",
			rollout: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"canary": map[string]interface{}{
							"currentStepAnalysisRuns": []interface{}{
								map[string]interface{}{
									"name":   "step-analysis-456",
									"status": "Successful",
								},
							},
						},
					},
				},
			},
			expected: []string{"step-analysis-456:Successful"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCurrentAnalysisRuns(tt.rollout)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsOwnedByRollout(t *testing.T) {
	tests := []struct {
		name        string
		run         *unstructured.Unstructured
		rolloutName string
		expected    bool
	}{
		{
			name:        "no owner references",
			run:         &unstructured.Unstructured{},
			rolloutName: "test-rollout",
			expected:    false,
		},
		{
			name: "owned by rollout",
			run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"kind": "Rollout",
								"name": "test-rollout",
							},
						},
					},
				},
			},
			rolloutName: "test-rollout",
			expected:    true,
		},
		{
			name: "owned by different rollout",
			run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"kind": "Rollout",
								"name": "other-rollout",
							},
						},
					},
				},
			},
			rolloutName: "test-rollout",
			expected:    false,
		},
		{
			name: "owned by different resource type",
			run: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"kind": "Deployment",
								"name": "test-rollout",
							},
						},
					},
				},
			},
			rolloutName: "test-rollout",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOwnedByRollout(tt.run, tt.rolloutName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
