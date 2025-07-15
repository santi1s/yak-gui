package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ArgoApp represents an ArgoCD application for the frontend
type ArgoApp struct {
	AppName    string   `json:"AppName"`
	Health     string   `json:"Health"`
	Sync       string   `json:"Sync"`
	Suspended  bool     `json:"Suspended"`
	SyncLoop   string   `json:"SyncLoop"`
	Conditions []string `json:"Conditions"`
}

// ArgoAppDetail represents detailed ArgoCD application info
type ArgoAppDetail struct {
	ArgoApp
	Namespace    string            `json:"namespace"`
	Project      string            `json:"project"`
	RepoURL      string            `json:"repoUrl"`
	Path         string            `json:"path"`
	TargetRev    string            `json:"targetRev"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	CreatedAt    string            `json:"createdAt"`
	Server       string            `json:"server"`
	Cluster      string            `json:"cluster"`
}

// ArgoResource represents an ArgoCD resource
type ArgoResource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Group     string `json:"group"`
	Namespace string `json:"namespace"`
	Health    string `json:"health"`
	Status    string `json:"status"`
	Orphaned  bool   `json:"orphaned"`
}

// ArgoConfig represents ArgoCD connection configuration
type ArgoConfig struct {
	Server   string `json:"server"`
	Project  string `json:"project"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// GetArgoApps retrieves ArgoCD applications
func (a *App) GetArgoApps(config ArgoConfig) ([]ArgoApp, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("ArgoCD server is required")
	}

	// Build yak command
	args := []string{"argocd", "status", "--json"}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		args = append(args, "--project", config.Project)
	}

	// Execute yak argocd status --json with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak argocd status failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yak argocd status timed out after 30 seconds")
		}
		return nil, fmt.Errorf("failed to execute yak argocd status: %w", err)
	}

	// Check if output looks like HTML (SAML redirect)
	outputStr := string(output)
	if strings.Contains(strings.ToLower(outputStr), "<!doctype") || 
	   strings.Contains(strings.ToLower(outputStr), "<html") ||
	   strings.Contains(strings.ToLower(outputStr), "saml") {
		return nil, fmt.Errorf("authentication required: received SAML redirect instead of JSON data. Please authenticate with ArgoCD first using 'yak argocd login' or try refreshing")
	}

	// Parse JSON output from yak argocd status
	var statusData map[string]interface{}
	if err := json.Unmarshal(output, &statusData); err != nil {
		return nil, fmt.Errorf("failed to parse yak argocd status output: %w", err)
	}

	// Convert to GUI format
	var apps []ArgoApp
	for appName, appDataInterface := range statusData {
		if appData, ok := appDataInterface.(map[string]interface{}); ok {
			app := ArgoApp{
				AppName:    appName,
				Health:     getString(appData, "Health"),
				Sync:       getString(appData, "Sync"),
				Suspended:  getBool(appData, "Suspended"),
				SyncLoop:   getString(appData, "SyncLoop"),
				Conditions: getStringSlice(appData, "Conditions"),
			}
			apps = append(apps, app)
		}
	}

	return apps, nil
}

// GetArgoAppDetail gets detailed information for a specific ArgoCD application
func (a *App) GetArgoAppDetail(config ArgoConfig, appName string) (*ArgoAppDetail, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("ArgoCD server is required")
	}
	if appName == "" {
		return nil, fmt.Errorf("application name is required")
	}

	// First, get the basic app status from yak argocd status
	statusArgs := []string{"argocd", "status", "-a", appName, "--json"}
	if config.Server != "" {
		statusArgs = append(statusArgs, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		statusArgs = append(statusArgs, "--project", config.Project)
	}


	// Execute yak argocd status --json with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	statusCmd := exec.CommandContext(ctx, findYakExecutable(), statusArgs...)
	statusOutput, err := statusCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak argocd status failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yak argocd status timed out after 30 seconds")
		}
		return nil, fmt.Errorf("failed to execute yak argocd status: %w", err)
	}


	// Parse JSON output from yak argocd status
	var statusData map[string]interface{}
	if err := json.Unmarshal(statusOutput, &statusData); err != nil {
		return nil, fmt.Errorf("failed to parse yak argocd status output: %w", err)
	}

	// Extract app status from the nested structure
	var appStatus ArgoApp
	if appData, ok := statusData[appName].(map[string]interface{}); ok {
		appStatus = ArgoApp{
			AppName:    getString(appData, "AppName"),
			Health:     getString(appData, "Health"),
			Sync:       getString(appData, "Sync"),
			Suspended:  getBool(appData, "Suspended"),
			SyncLoop:   getString(appData, "SyncLoop"),
			Conditions: getStringSlice(appData, "Conditions"),
		}
	} else {
		return nil, fmt.Errorf("app %s not found in status response", appName)
	}

	// Second, get detailed configuration from yak argocd get
	getArgs := []string{"argocd", "get", "-a", appName, "--json"}
	if config.Server != "" {
		getArgs = append(getArgs, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		getArgs = append(getArgs, "--project", config.Project)
	}


	getCmd := exec.CommandContext(ctx, findYakExecutable(), getArgs...)
	getOutput, err := getCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak argocd get failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yak argocd get timed out after 30 seconds")
		}
		return nil, fmt.Errorf("failed to execute yak argocd get: %w", err)
	}


	// Parse JSON output from yak argocd get
	var appDetailData map[string]interface{}
	if err := json.Unmarshal(getOutput, &appDetailData); err != nil {
		return nil, fmt.Errorf("failed to parse yak argocd get output: %w", err)
	}

	// Combine both sources of data
	appDetail := &ArgoAppDetail{
		ArgoApp:     appStatus, // Use the status data for basic app info
		Namespace:   getNestedString(appDetailData, "spec", "destination", "namespace"),
		Project:     getNestedString(appDetailData, "spec", "project"),
		RepoURL:     getNestedString(appDetailData, "spec", "source", "repoURL"),
		Path:        getNestedString(appDetailData, "spec", "source", "path"),
		TargetRev:   getNestedString(appDetailData, "spec", "source", "targetRevision"),
		CreatedAt:   getNestedString(appDetailData, "metadata", "creationTimestamp"),
		Server:      config.Server,
		Cluster:     getNestedString(appDetailData, "spec", "destination", "server"),
	}

	// Handle labels and annotations from metadata
	if metadata, ok := appDetailData["metadata"].(map[string]interface{}); ok {
		if labels, ok := metadata["labels"].(map[string]interface{}); ok {
			appDetail.Labels = make(map[string]string)
			for k, v := range labels {
				if strVal, ok := v.(string); ok {
					appDetail.Labels[k] = strVal
				}
			}
		}

		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			appDetail.Annotations = make(map[string]string)
			for k, v := range annotations {
				if strVal, ok := v.(string); ok {
					appDetail.Annotations[k] = strVal
				}
			}
		}
	}

	// Ensure AppName is set
	if appDetail.ArgoApp.AppName == "" {
		appDetail.ArgoApp.AppName = appName
	}

	return appDetail, nil
}

// SyncArgoApp synchronizes an ArgoCD application
func (a *App) SyncArgoApp(config ArgoConfig, appName string, prune, dryRun bool) error {
	// Build yak command
	args := []string{"argocd", "sync", "-a", appName}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		args = append(args, "--project", config.Project)
	}
	if prune {
		args = append(args, "--prune")
	}
	if dryRun {
		args = append(args, "--dry-run")
	}

	// Execute yak argocd sync
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to sync ArgoCD app: %w", err)
	}

	return nil
}

// RefreshArgoApp refreshes an ArgoCD application
func (a *App) RefreshArgoApp(config ArgoConfig, appName string) error {
	// Build yak command
	args := []string{"argocd", "refresh", "-a", appName}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		args = append(args, "--project", config.Project)
	}

	// Execute yak argocd refresh
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to refresh ArgoCD app: %w", err)
	}

	return nil
}

// SuspendArgoApp suspends an ArgoCD application
func (a *App) SuspendArgoApp(config ArgoConfig, appName string) error {
	// Build yak command
	args := []string{"argocd", "suspend", "-a", appName}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		args = append(args, "--project", config.Project)
	}

	// Execute yak argocd suspend
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to suspend ArgoCD app: %w", err)
	}

	return nil
}

// UnsuspendArgoApp unsuspends an ArgoCD application
func (a *App) UnsuspendArgoApp(config ArgoConfig, appName string) error {
	// Build yak command
	args := []string{"argocd", "unsuspend", "-a", appName}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}
	if config.Project != "" {
		args = append(args, "--project", config.Project)
	}

	// Execute yak argocd unsuspend
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unsuspend ArgoCD app: %w", err)
	}

	return nil
}

// LoginToArgoCD logs in to ArgoCD
func (a *App) LoginToArgoCD(config ArgoConfig) error {
	// Build yak command
	args := []string{"argocd", "login"}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}

	// Execute yak argocd login
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to login to ArgoCD: %w", err)
	}

	return nil
}

// GetArgoCDServerFromProfile gets the ArgoCD server from AWS profile
func (a *App) GetArgoCDServerFromProfile() (string, error) {
	// Get current AWS profile
	profile := a.GetCurrentAWSProfile()
	if profile == "" {
		return "", fmt.Errorf("AWS_PROFILE environment variable is not set")
	}

	// Remove '-sso' suffix if present
	cleanProfile := strings.TrimSuffix(profile, "-sso")
	
	// Construct ArgoCD server URL
	server := fmt.Sprintf("argocd-%s.doctolib.net", cleanProfile)
	
	return server, nil
}