package argocd

import (
	"testing"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {
	// Test command structure
	assert.NotNil(t, getCmd)
	assert.Equal(t, "get", getCmd.Use)
	assert.Equal(t, "Get detailed information about an ArgoCD application", getCmd.Short)
}

func TestGetCommandFlags(t *testing.T) {
	// Test that application flag exists and is required
	appFlag := getCmd.Flags().Lookup("application")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)
}

func TestFormatSource(t *testing.T) {
	tests := []struct {
		name     string
		source   *v1alpha1.ApplicationSource
		expected string
	}{
		{
			name:     "nil source",
			source:   nil,
			expected: "<none>",
		},
		{
			name: "complete source",
			source: &v1alpha1.ApplicationSource{
				RepoURL:        "https://github.com/example/repo.git",
				Path:           "manifests/app",
				TargetRevision: "main",
			},
			expected: "repo:repo\npath:manifests/app\nrev:main",
		},
		{
			name: "helm chart source",
			source: &v1alpha1.ApplicationSource{
				RepoURL:        "https://charts.example.com",
				Chart:          "my-chart",
				TargetRevision: "1.0.0",
			},
			expected: "repo:charts.example.com\nrev:1.0.0\nchart:my-chart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSource(tt.source)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDestination(t *testing.T) {
	tests := []struct {
		name        string
		destination v1alpha1.ApplicationDestination
		expected    string
	}{
		{
			name: "in-cluster destination",
			destination: v1alpha1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "default",
			},
			expected: "server:in-cluster\nnamespace:default",
		},
		{
			name: "external cluster",
			destination: v1alpha1.ApplicationDestination{
				Server:    "https://external-cluster.example.com",
				Namespace: "production",
			},
			expected: "server:https://external-cluster.example.com\nnamespace:production",
		},
		{
			name: "cluster by name",
			destination: v1alpha1.ApplicationDestination{
				Name:      "staging-cluster",
				Namespace: "staging",
			},
			expected: "cluster:staging-cluster\nnamespace:staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDestination(tt.destination)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatSyncPolicy(t *testing.T) {
	tests := []struct {
		name     string
		policy   *v1alpha1.SyncPolicy
		expected string
	}{
		{
			name:     "nil policy",
			policy:   nil,
			expected: "<none>",
		},
		{
			name: "manual sync",
			policy: &v1alpha1.SyncPolicy{
				SyncOptions: []string{"CreateNamespace=true"},
			},
			expected: "manual\nopts:CreateNamespace=true",
		},
		{
			name: "automated sync with prune and self-heal",
			policy: &v1alpha1.SyncPolicy{
				Automated: &v1alpha1.SyncPolicyAutomated{
					Prune:    true,
					SelfHeal: true,
				},
				SyncOptions: []string{"CreateNamespace=true", "PruneLast=true"},
			},
			expected: "auto:prune,self-heal\nopts:CreateNamespace=true,PruneLast=true",
		},
		{
			name: "automated sync without options",
			policy: &v1alpha1.SyncPolicy{
				Automated: &v1alpha1.SyncPolicyAutomated{
					Prune:    false,
					SelfHeal: false,
				},
			},
			expected: "auto:enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSyncPolicy(tt.policy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatIgnoredDifferences(t *testing.T) {
	tests := []struct {
		name     string
		diffs    []v1alpha1.ResourceIgnoreDifferences
		expected string
	}{
		{
			name:     "no ignored differences",
			diffs:    []v1alpha1.ResourceIgnoreDifferences{},
			expected: "<none>",
		},
		{
			name: "single ignored difference with JSON pointers",
			diffs: []v1alpha1.ResourceIgnoreDifferences{
				{
					Group:        "apps",
					Kind:         "Deployment",
					Name:         "my-app",
					JSONPointers: []string{"/metadata/generation", "/spec/replicas"},
				},
			},
			expected: "apps/Deployment:my-app\n  Fields: /metadata/generation, /spec/replicas",
		},
		{
			name: "ignored difference with JQ expressions",
			diffs: []v1alpha1.ResourceIgnoreDifferences{
				{
					Group:             "apps",
					Kind:              "Deployment",
					JQPathExpressions: []string{".spec.template.spec.containers[].image"},
				},
			},
			expected: "apps/Deployment\n  JQ: .spec.template.spec.containers[].image",
		},
		{
			name: "multiple ignored differences",
			diffs: []v1alpha1.ResourceIgnoreDifferences{
				{
					Group:        "apps",
					Kind:         "Deployment",
					JSONPointers: []string{"/metadata/generation"},
				},
				{
					Group:                 "",
					Kind:                  "ConfigMap",
					Name:                  "my-config",
					ManagedFieldsManagers: []string{"kubectl"},
				},
			},
			expected: "apps/Deployment\n  Fields: /metadata/generation\n---\nConfigMap:my-config\n  Managers: kubectl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatIgnoredDifferences(tt.diffs)
			assert.Equal(t, tt.expected, result)
		})
	}
}
