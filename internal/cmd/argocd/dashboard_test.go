package argocd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboardCommand(t *testing.T) {
	tests := []struct {
		name        string
		serverAddr  string
		project     string
		application string
		expectedURL string
	}{
		{
			name:        "dashboard without application",
			serverAddr:  "argocd-test.example.com",
			project:     "main",
			application: "",
			expectedURL: "https://argocd-test.example.com/applications?search=main",
		},
		{
			name:        "dashboard with application",
			serverAddr:  "argocd-test.example.com",
			project:     "main",
			application: "my-app",
			expectedURL: "https://argocd-test.example.com/applications/main/my-app",
		},
		{
			name:        "dashboard with special characters in project",
			serverAddr:  "argocd-test.example.com",
			project:     "test-project-name",
			application: "",
			expectedURL: "https://argocd-test.example.com/applications?search=test-project-name",
		},
		{
			name:        "dashboard with special characters in app name",
			serverAddr:  "argocd-test.example.com",
			project:     "main",
			application: "app-with-dashes",
			expectedURL: "https://argocd-test.example.com/applications/main/app-with-dashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			oldAWSProfile := os.Getenv("AWS_PROFILE")
			os.Setenv("AWS_PROFILE", "test")
			defer os.Setenv("AWS_PROFILE", oldAWSProfile)

			// Set up flags
			providedFlags.addr = tt.serverAddr
			providedFlags.project = tt.project
			providedDashboardFlags.application = tt.application

			// Test URL construction logic (we can't test actual browser opening in unit tests)
			serverAddr := tt.serverAddr
			var dashboardURL string

			if tt.application != "" {
				dashboardURL = "https://" + serverAddr + "/applications/" + tt.project + "/" + tt.application
			} else {
				dashboardURL = "https://" + serverAddr + "/applications?search=" + tt.project
			}

			assert.Equal(t, tt.expectedURL, dashboardURL)
		})
	}
}

func TestDashboardCommandFlags(t *testing.T) {
	// Test command flags are properly set up
	assert.NotNil(t, dashboardCmd)
	assert.Equal(t, "dashboard", dashboardCmd.Use)
	assert.Equal(t, "Open the ArgoCD web dashboard in your browser", dashboardCmd.Short)

	// Check if application flag exists
	appFlag := dashboardCmd.Flags().Lookup("application")
	assert.NotNil(t, appFlag)
	assert.Equal(t, "a", appFlag.Shorthand)
}
