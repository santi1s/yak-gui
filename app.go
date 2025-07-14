package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// findYakExecutable searches for yak executable in common paths
func findYakExecutable() string {
	// Common paths where yak might be installed
	paths := []string{
		"/opt/homebrew/bin/yak",    // Homebrew on Apple Silicon
		"/usr/local/bin/yak",       // Homebrew on Intel Macs
		"/usr/bin/yak",             // System installation
		"yak",                      // Try PATH first
	}
	
	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}
	return "yak" // fallback to PATH
}

// Simple test function to see if Wails generates any bindings
func Greet(name string) string {
	return fmt.Sprintf("Hello %s from Wails!", name)
}

// App struct - Wails app context
type App struct {
	ctx context.Context
}

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

// RolloutStatus represents an Argo Rollout status
type RolloutStatus struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Status      string            `json:"status"`
	Replicas    string            `json:"replicas"`
	Updated     string            `json:"updated"`
	Ready       string            `json:"ready"`
	Available   string            `json:"available"`
	Strategy    string            `json:"strategy"`
	CurrentStep string            `json:"currentStep"`
	Revision    string            `json:"revision"`
	Message     string            `json:"message"`
	Analysis    string            `json:"analysis"`
	Images      map[string]string `json:"images"`
}

// RolloutListItem represents a rollout in list view
type RolloutListItem struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Status    string            `json:"status"`
	Replicas  string            `json:"replicas"`
	Age       string            `json:"age"`
	Strategy  string            `json:"strategy"`
	Revision  string            `json:"revision"`
	Images    map[string]string `json:"images"`
}

// KubernetesConfig represents Kubernetes connection configuration
type KubernetesConfig struct {
	Server    string `json:"server"`
	Namespace string `json:"namespace"`
}

// SecretConfig represents secret management configuration
type SecretConfig struct {
	Platform    string `json:"platform"`
	Environment string `json:"environment"`
	Team        string `json:"team"`
}

// SecretListItem represents a secret in list view
type SecretListItem struct {
	Path        string `json:"path"`
	Version     int    `json:"version"`
	Owner       string `json:"owner"`
	Usage       string `json:"usage"`
	Source      string `json:"source"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// SecretData represents secret key-value data
type SecretData struct {
	Path     string            `json:"path"`
	Version  int               `json:"version"`
	Data     map[string]string `json:"data"`
	Metadata SecretMetadata    `json:"metadata"`
}

// SecretMetadata represents secret metadata
type SecretMetadata struct {
	Owner       string `json:"owner"`
	Usage       string `json:"usage"`
	Source      string `json:"source"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Version     int    `json:"version"`
	Destroyed   bool   `json:"destroyed"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts, before the frontend is loaded
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add any logic here that needs to run after the DOM is ready
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// Return true to prevent the application from quitting
	return false
}

// shutdown is called during application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform any teardown of resources here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// TestSimpleArray returns a simple array to test Wails binding
func (a *App) TestSimpleArray() []string {
	return []string{"app1", "app2", "app3"}
}

// GetArgoCDServerFromProfile gets the ArgoCD server address from AWS_PROFILE environment variable
func (a *App) GetArgoCDServerFromProfile() (string, error) {
	awsProfile := os.Getenv("AWS_PROFILE")
	if awsProfile == "" {
		return "", fmt.Errorf("AWS_PROFILE environment variable is not set")
	}
	return "argocd-" + awsProfile + ".doctolib.net", nil
}

// GetCurrentAWSProfile returns the current AWS_PROFILE environment variable
func (a *App) GetCurrentAWSProfile() string {
	return os.Getenv("AWS_PROFILE")
}

// SetAWSProfile sets the AWS_PROFILE environment variable and auto-configures KUBECONFIG and kubectl context
func (a *App) SetAWSProfile(profile string) error {
	if profile == "" {
		return fmt.Errorf("AWS profile cannot be empty")
	}
	
	// Set AWS_PROFILE
	if err := os.Setenv("AWS_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set AWS_PROFILE: %v", err)
	}
	
	// Auto-generate KUBECONFIG path if TFINFRA_REPOSITORY_PATH is available
	tfRepoPath := os.Getenv("TFINFRA_REPOSITORY_PATH")
	if tfRepoPath != "" {
		kubeconfigPath := filepath.Join(tfRepoPath, "setup", "k8senv", profile, "config")
		
		// Check if the kubeconfig file exists
		if _, err := os.Stat(kubeconfigPath); err == nil {
			// Set KUBECONFIG
			if err := os.Setenv("KUBECONFIG", kubeconfigPath); err != nil {
				return fmt.Errorf("failed to set KUBECONFIG: %v", err)
			}
			
			// Set kubectl context to match the profile
			cmd := exec.Command("kubectl", "config", "use-context", profile)
			cmd.Env = os.Environ() // Use current environment including updated KUBECONFIG
			
			if output, err := cmd.CombinedOutput(); err != nil {
				// Don't fail the whole operation if kubectl context setting fails
				fmt.Printf("Warning: failed to set kubectl context to %s: %v\nOutput: %s\n", profile, err, string(output))
			}
		} else {
			fmt.Printf("Warning: kubeconfig file not found at %s\n", kubeconfigPath)
		}
	} else {
		fmt.Println("Warning: TFINFRA_REPOSITORY_PATH environment variable not set, cannot auto-configure KUBECONFIG")
	}
	
	return nil
}

// GetKubeconfig returns the current KUBECONFIG environment variable
func (a *App) GetKubeconfig() string {
	return os.Getenv("KUBECONFIG")
}

// SetKubeconfig sets the KUBECONFIG environment variable for the current session
func (a *App) SetKubeconfig(path string) error {
	if path == "" {
		return fmt.Errorf("Kubeconfig path cannot be empty")
	}
	return os.Setenv("KUBECONFIG", path)
}

// SetPATH sets the PATH environment variable for the current session
func (a *App) SetPATH(path string) error {
	if path == "" {
		return fmt.Errorf("PATH cannot be empty")
	}
	return os.Setenv("PATH", path)
}

// SetTfInfraRepositoryPath sets the TFINFRA_REPOSITORY_PATH environment variable for the current session
func (a *App) SetTfInfraRepositoryPath(path string) error {
	if path == "" {
		return fmt.Errorf("TFINFRA_REPOSITORY_PATH cannot be empty")
	}
	return os.Setenv("TFINFRA_REPOSITORY_PATH", path)
}

// GetAWSProfiles reads ~/.aws/config and returns available profiles (excluding -sso profiles)
func (a *App) GetAWSProfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	configPath := filepath.Join(homeDir, ".aws", "config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // Return empty list if config doesn't exist
		}
		return nil, fmt.Errorf("failed to open AWS config file: %v", err)
	}
	defer file.Close()
	
	var profiles []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Look for profile sections: [profile profile-name] or [default]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			line = strings.Trim(line, "[]")
			
			var profileName string
			if line == "default" {
				profileName = "default"
			} else if strings.HasPrefix(line, "profile ") {
				profileName = strings.TrimPrefix(line, "profile ")
			} else {
				continue // Skip non-profile sections
			}
			
			// Exclude profiles ending with -sso
			if !strings.HasSuffix(profileName, "-sso") {
				profiles = append(profiles, profileName)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading AWS config file: %v", err)
	}
	
	// Sort profiles for consistent ordering
	sort.Strings(profiles)
	return profiles, nil
}

// GetShellPATH attempts to get PATH from common shell configuration files
func (a *App) GetShellPATH() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	// Common shell config files to check (in order of preference)
	configFiles := []string{
		".zshrc",
		".bashrc", 
		".bash_profile",
		".profile",
	}
	
	// Try to detect shell from environment
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		// For zsh, also check .zprofile
		configFiles = append([]string{".zshrc", ".zprofile"}, configFiles[1:]...)
	}
	
	// Look for PATH exports in shell config files
	for _, configFile := range configFiles {
		configPath := filepath.Join(homeDir, configFile)
		file, err := os.Open(configPath)
		if err != nil {
			continue // Skip if file doesn't exist
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			
			// Look for PATH exports (various formats)
			if strings.HasPrefix(line, "export PATH=") {
				path := strings.TrimPrefix(line, "export PATH=")
				path = strings.Trim(path, "\"'") // Remove quotes
				return path, nil
			} else if strings.HasPrefix(line, "PATH=") {
				path := strings.TrimPrefix(line, "PATH=")
				path = strings.Trim(path, "\"'") // Remove quotes
				return path, nil
			}
		}
	}
	
	// Fallback to common macOS paths if we can't find it in config files
	commonPaths := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin", 
		"/usr/bin",
		"/bin",
		"/usr/sbin",
		"/sbin",
	}
	
	return strings.Join(commonPaths, ":"), nil
}

// GetEnvironmentVariables returns a map of current environment variables
func (a *App) GetEnvironmentVariables() map[string]string {
	return map[string]string{
		"AWS_PROFILE":              os.Getenv("AWS_PROFILE"),
		"KUBECONFIG":               os.Getenv("KUBECONFIG"),
		"HOME":                     os.Getenv("HOME"),
		"PATH":                     os.Getenv("PATH"),
		"TFINFRA_REPOSITORY_PATH":  os.Getenv("TFINFRA_REPOSITORY_PATH"),
	}
}

// LoginToArgoCD attempts to login to ArgoCD using the yak CLI
func (a *App) LoginToArgoCD(config ArgoConfig) error {
	if config.Server == "" {
		return fmt.Errorf("ArgoCD server is required")
	}

	// Build yak login command
	args := []string{"argocd", "login"}
	if config.Server != "" {
		args = append(args, "--argocd-addr", config.Server)
	}

	// Execute yak argocd login with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	
	// Run the command and wait for completion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to login to ArgoCD: %w", err)
	}

	return nil
}


// GetArgoApps gets all ArgoCD applications for a project by calling the yak CLI
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
	
	// Add debugging
	
	output, err := cmd.Output()
	if err != nil {
		// Get more detailed error information
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak command failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yak command timed out after 30 seconds")
		}
		return nil, fmt.Errorf("failed to execute yak command: %w", err)
	}
	

	// Check if output looks like HTML (SAML redirect)
	outputStr := string(output)
	if strings.Contains(strings.ToLower(outputStr), "<!doctype") || 
	   strings.Contains(strings.ToLower(outputStr), "<html") ||
	   strings.Contains(strings.ToLower(outputStr), "saml") {
		return nil, fmt.Errorf("authentication required: received SAML redirect instead of JSON data. Please authenticate with ArgoCD first using 'yak argocd login' or try refreshing")
	}

	// Parse JSON output from your CLI
	var statusData map[string]interface{}
	if err := json.Unmarshal(output, &statusData); err != nil {
		return nil, fmt.Errorf("failed to parse yak output: %w", err)
	}
	

	// Convert to GUI format
	var apps []ArgoApp
	for appName, appDataInterface := range statusData {
		appData, ok := appDataInterface.(map[string]interface{})
		if !ok {
			continue
		}

		app := ArgoApp{
			AppName:    getString(appData, "AppName"),
			Health:     getString(appData, "Health"),
			Sync:       getString(appData, "Sync"),
			Suspended:  getBool(appData, "Suspended"),
			SyncLoop:   getString(appData, "SyncLoop"),
			Conditions: getStringSlice(appData, "Conditions"),
		}
		
		// If AppName field is empty, use the map key as the name
		if app.AppName == "" {
			app.AppName = appName
		}
		
		apps = append(apps, app)
	}


	// Sort by name for consistent ordering
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].AppName < apps[j].AppName
	})

	return apps, nil
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
		return fmt.Errorf("failed to sync application %s: %w", appName, err)
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
		return fmt.Errorf("failed to refresh application %s: %w", appName, err)
	}

	return nil
}

// SuspendArgoApp suspends an ArgoCD application using sync windows
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
		return fmt.Errorf("failed to suspend application %s: %w", appName, err)
	}

	return nil
}

// UnsuspendArgoApp removes suspension from an ArgoCD application
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
		return fmt.Errorf("failed to unsuspend application %s: %w", appName, err)
	}

	return nil
}

// GetRollouts gets all rollouts using the yak CLI
func (a *App) GetRollouts(config KubernetesConfig) ([]RolloutListItem, error) {
	// Build yak command
	args := []string{"rollouts", "list", "--json"}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "--namespace", config.Namespace)
	} else {
		args = append(args, "--all")
	}

	// Execute yak rollouts list --json with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak rollouts list failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yak rollouts list timed out after 30 seconds")
		}
		return nil, fmt.Errorf("failed to execute yak rollouts list: %w", err)
	}
	

	// Check if output looks like HTML (authentication issues)
	outputStr := string(output)
	if strings.Contains(strings.ToLower(outputStr), "<!doctype") || 
	   strings.Contains(strings.ToLower(outputStr), "<html") {
		return nil, fmt.Errorf("authentication required: received HTML response instead of JSON data")
	}

	// Parse JSON output - yak rollouts returns a Kubernetes List object
	var listResponse struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(output, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse yak rollouts output: %w", err)
	}

	// Convert Kubernetes rollout objects to our simplified structure
	var rollouts []RolloutListItem
	for _, item := range listResponse.Items {
		metadata, ok := item["metadata"].(map[string]interface{})
		if !ok {
			continue
		}
		
		spec, _ := item["spec"].(map[string]interface{})
		status, _ := item["status"].(map[string]interface{})
		
		rollout := RolloutListItem{
			Name:      getString(metadata, "name"),
			Namespace: getString(metadata, "namespace"),
			Status:    getRolloutPhase(status),
			Replicas:  getRolloutReplicas(status),
			Age:       getRolloutAge(metadata),
			Strategy:  getRolloutStrategy(spec),
			Revision:  getRolloutRevision(metadata),
			Images:    getRolloutImages(spec),
		}
		
		rollouts = append(rollouts, rollout)
	}

	return rollouts, nil
}

// GetRolloutStatus gets detailed status for a specific rollout
func (a *App) GetRolloutStatus(config KubernetesConfig, rolloutName string) (*RolloutStatus, error) {
	if rolloutName == "" {
		return nil, fmt.Errorf("rollout name is required")
	}

	// Build yak command - use 'get' instead of 'status' to get JSON object
	args := []string{"rollouts", "get", "-r", rolloutName, "--json"}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "-n", config.Namespace)
	}

	// Execute yak rollouts get with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak rollouts get failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute yak rollouts get: %w", err)
	}

	// Debug: log the raw output to understand what we're getting
	fmt.Printf("DEBUG: yak rollouts get raw output: %s\n", string(output))
	
	// Parse JSON output - expecting a Kubernetes object
	var rolloutObj map[string]interface{}
	if err := json.Unmarshal(output, &rolloutObj); err != nil {
		// Include the raw output in the error for debugging
		outputPreview := string(output)
		if len(outputPreview) > 200 {
			outputPreview = outputPreview[:200] + "..."
		}
		return nil, fmt.Errorf("failed to parse rollout status (raw output: %s): %w", outputPreview, err)
	}

	// Extract metadata, spec, and status
	metadata, _ := rolloutObj["metadata"].(map[string]interface{})
	spec, _ := rolloutObj["spec"].(map[string]interface{})
	statusObj, _ := rolloutObj["status"].(map[string]interface{})

	// Build RolloutStatus from Kubernetes object
	status := &RolloutStatus{
		Name:        getString(metadata, "name"),
		Namespace:   getString(metadata, "namespace"),
		Status:      getRolloutPhase(statusObj),
		Replicas:    getRolloutReplicas(statusObj),
		Updated:     "0", // TODO: extract from status if available
		Ready:       "0", // TODO: extract from status if available
		Available:   "0", // TODO: extract from status if available
		Strategy:    getRolloutStrategy(spec),
		CurrentStep: "0", // TODO: extract from status if available
		Revision:    getRolloutRevision(metadata),
		Message:     getString(statusObj, "message"),
		Analysis:    "", // TODO: extract analysis if available
		Images:      getRolloutImages(spec),
	}

	return status, nil
}

// PromoteRollout promotes a rollout to the next step or full deployment
func (a *App) PromoteRollout(config KubernetesConfig, rolloutName string, full bool) error {
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	// Build yak command
	args := []string{"rollouts", "promote", "-r", rolloutName}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "--namespace", config.Namespace)
	}
	if full {
		args = append(args, "--full")
	}

	// Execute yak rollouts promote
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to promote rollout %s: %w", rolloutName, err)
	}

	return nil
}

// PauseRollout pauses a rollout
func (a *App) PauseRollout(config KubernetesConfig, rolloutName string) error {
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	// Build yak command
	args := []string{"rollouts", "pause", "-r", rolloutName}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "--namespace", config.Namespace)
	}

	// Execute yak rollouts pause
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pause rollout %s: %w", rolloutName, err)
	}

	return nil
}

// AbortRollout aborts a rollout and rolls back to stable version
func (a *App) AbortRollout(config KubernetesConfig, rolloutName string) error {
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	// Build yak command
	args := []string{"rollouts", "abort", "-r", rolloutName}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "--namespace", config.Namespace)
	}

	// Execute yak rollouts abort
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to abort rollout %s: %w", rolloutName, err)
	}

	return nil
}

// RestartRollout restarts rollout pods
func (a *App) RestartRollout(config KubernetesConfig, rolloutName string) error {
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}

	// Build yak command
	args := []string{"rollouts", "restart", "-r", rolloutName}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "--namespace", config.Namespace)
	}

	// Execute yak rollouts restart
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart rollout %s: %w", rolloutName, err)
	}

	return nil
}

// SetRolloutImage updates the image for a rollout
func (a *App) SetRolloutImage(config KubernetesConfig, rolloutName, image, container string) error {
	if rolloutName == "" {
		return fmt.Errorf("rollout name is required")
	}
	if image == "" {
		return fmt.Errorf("image is required")
	}

	// Build yak command
	args := []string{"rollouts", "set-image", "-r", rolloutName, "--image", image}
	if config.Server != "" {
		args = append(args, "--server", config.Server)
	}
	if config.Namespace != "" {
		args = append(args, "--namespace", config.Namespace)
	}
	if container != "" {
		args = append(args, "--container", container)
	}

	// Execute yak rollouts set-image
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set image for rollout %s: %w", rolloutName, err)
	}

	return nil
}

// GetSecrets lists secrets from a path using the yak CLI
func (a *App) GetSecrets(config SecretConfig, path string) ([]SecretListItem, error) {
	// Build yak command
	args := []string{"secret", "list", "--json"}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}
	if path != "" {
		args = append(args, "--path", path)
	}

	// Execute yak secret list --json with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak secret list failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yak secret list timed out after 30 seconds")
		}
		return nil, fmt.Errorf("failed to execute yak secret list: %w", err)
	}
	

	// Parse JSON output - yak secret may return various formats
	// First try to parse as an array directly
	var secrets []SecretListItem
	if err := json.Unmarshal(output, &secrets); err != nil {
		// If that fails, try to parse as a map/object
		var secretMap map[string]interface{}
		if mapErr := json.Unmarshal(output, &secretMap); mapErr != nil {
			return nil, fmt.Errorf("failed to parse yak secret output: %w", err)
		}
		
		// If it's a map, try to extract secrets from it
		secrets = parseSecretsFromMap(secretMap)
	}

	return secrets, nil
}

// GetSecretData retrieves secret data from a specific path
func (a *App) GetSecretData(config SecretConfig, path string, version int) (*SecretData, error) {
	if path == "" {
		return nil, fmt.Errorf("secret path is required")
	}

	// Build yak command
	args := []string{"secret", "get", "--json", "--path", path}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}
	if version > 0 {
		args = append(args, "--version", fmt.Sprintf("%d", version))
	}

	// Execute yak secret get with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yak secret get failed with exit code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute yak secret get: %w", err)
	}

	// Parse JSON output
	var secretData SecretData
	if err := json.Unmarshal(output, &secretData); err != nil {
		return nil, fmt.Errorf("failed to parse secret data: %w", err)
	}

	return &secretData, nil
}

// CreateSecret creates a new secret
func (a *App) CreateSecret(config SecretConfig, path, owner, usage, source string, data map[string]string) error {
	if path == "" {
		return fmt.Errorf("secret path is required")
	}
	if owner == "" {
		return fmt.Errorf("owner is required")
	}
	if usage == "" {
		return fmt.Errorf("usage is required")
	}
	if source == "" {
		return fmt.Errorf("source is required")
	}

	// Build yak command
	args := []string{"secret", "create", "--path", path, "--owner", owner, "--usage", usage, "--source", source}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}

	// Add data as key=value pairs
	for key, value := range data {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute yak secret create
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create secret %s: %w", path, err)
	}

	return nil
}

// UpdateSecret updates an existing secret
func (a *App) UpdateSecret(config SecretConfig, path string, data map[string]string) error {
	if path == "" {
		return fmt.Errorf("secret path is required")
	}

	// Build yak command
	args := []string{"secret", "update", "--path", path}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}

	// Add data as key=value pairs
	for key, value := range data {
		args = append(args, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute yak secret update
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update secret %s: %w", path, err)
	}

	return nil
}

// DeleteSecret deletes a secret version
func (a *App) DeleteSecret(config SecretConfig, path string, version int) error {
	if path == "" {
		return fmt.Errorf("secret path is required")
	}

	// Build yak command
	args := []string{"secret", "delete", "--path", path}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}
	if version > 0 {
		args = append(args, "--version", fmt.Sprintf("%d", version))
	}

	// Execute yak secret delete
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete secret %s: %w", path, err)
	}

	return nil
}

// parseSecretsFromMap extracts secrets from a map structure
func parseSecretsFromMap(secretMap map[string]interface{}) []SecretListItem {
	var secrets []SecretListItem
	
	// Check if it has a "keys" array (the actual format returned by yak secret list)
	if keysArray, ok := secretMap["keys"].([]interface{}); ok {
		for _, keyItem := range keysArray {
			if keyStr, ok := keyItem.(string); ok {
				// Create a basic SecretListItem with just the path
				// Since yak secret list only returns paths, we'll need metadata from individual gets
				secret := SecretListItem{
					Path:      keyStr,
					Version:   0,        // Unknown from list command
					Owner:     "Unknown", // Unknown from list command
					Usage:     "Unknown", // Unknown from list command
					Source:    "Unknown", // Unknown from list command
					CreatedAt: "",       // Unknown from list command
					UpdatedAt: "",       // Unknown from list command
				}
				secrets = append(secrets, secret)
			}
		}
		return secrets
	}
	
	// Try other possible structures
	if secretsArray, ok := secretMap["secrets"].([]interface{}); ok {
		secrets = parseSecretArray(secretsArray)
	} else if dataArray, ok := secretMap["data"].([]interface{}); ok {
		secrets = parseSecretArray(dataArray)
	} else if itemsArray, ok := secretMap["items"].([]interface{}); ok {
		secrets = parseSecretArray(itemsArray)
	} else {
		// If the entire map appears to be secrets, treat each key as a secret path
		for path, secretData := range secretMap {
			if secretInfo, ok := secretData.(map[string]interface{}); ok {
				secret := SecretListItem{
					Path:      path,
					Version:   getInt(secretInfo, "version"),
					Owner:     getString(secretInfo, "owner"),
					Usage:     getString(secretInfo, "usage"),
					Source:    getString(secretInfo, "source"),
					CreatedAt: getString(secretInfo, "created_at"),
					UpdatedAt: getString(secretInfo, "updated_at"),
				}
				secrets = append(secrets, secret)
			}
		}
	}
	
	return secrets
}

// parseSecretArray converts an array of secret objects to SecretListItem
func parseSecretArray(secretsArray []interface{}) []SecretListItem {
	var secrets []SecretListItem
	
	for _, item := range secretsArray {
		if secretInfo, ok := item.(map[string]interface{}); ok {
			secret := SecretListItem{
				Path:      getString(secretInfo, "path"),
				Version:   getInt(secretInfo, "version"),
				Owner:     getString(secretInfo, "owner"),
				Usage:     getString(secretInfo, "usage"),
				Source:    getString(secretInfo, "source"),
				CreatedAt: getString(secretInfo, "created_at"),
				UpdatedAt: getString(secretInfo, "updated_at"),
			}
			secrets = append(secrets, secret)
		}
	}
	
	return secrets
}

// getMapKeys returns the keys of a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper functions to safely extract values from map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getStringSlice(data map[string]interface{}, key string) []string {
	if val, ok := data[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

// Helper functions for rollout parsing
func getRolloutPhase(status map[string]interface{}) string {
	if status == nil {
		return "Unknown"
	}
	
	// Check for phase first
	if phase := getString(status, "phase"); phase != "" {
		return phase
	}
	
	// Fallback to health status
	if health := getString(status, "health"); health != "" {
		return health
	}
	
	return "Unknown"
}

func getRolloutReplicas(status map[string]interface{}) string {
	if status == nil {
		return "0/0"
	}
	
	replicas := getInt(status, "replicas")
	readyReplicas := getInt(status, "readyReplicas")
	
	return fmt.Sprintf("%d/%d", readyReplicas, replicas)
}

func getRolloutAge(metadata map[string]interface{}) string {
	if metadata == nil {
		return "Unknown"
	}
	
	creationTimestamp := getString(metadata, "creationTimestamp")
	if creationTimestamp == "" {
		return "Unknown"
	}
	
	// Parse the timestamp and calculate age
	if t, err := time.Parse(time.RFC3339, creationTimestamp); err == nil {
		duration := time.Since(t)
		if duration.Hours() > 24 {
			return fmt.Sprintf("%dd", int(duration.Hours()/24))
		} else if duration.Hours() > 1 {
			return fmt.Sprintf("%dh", int(duration.Hours()))
		} else {
			return fmt.Sprintf("%dm", int(duration.Minutes()))
		}
	}
	
	return "Unknown"
}

func getRolloutStrategy(spec map[string]interface{}) string {
	if spec == nil {
		return "Unknown"
	}
	
	strategy, ok := spec["strategy"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}
	
	if _, hasCanary := strategy["canary"]; hasCanary {
		return "Canary"
	}
	if _, hasBlueGreen := strategy["blueGreen"]; hasBlueGreen {
		return "BlueGreen"
	}
	
	return "Unknown"
}

func getRolloutRevision(metadata map[string]interface{}) string {
	if metadata == nil {
		return "0"
	}
	
	// Get annotations
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		return "0"
	}
	
	// Try to get revision from rollout.argoproj.io/revision annotation
	if revision := getString(annotations, "rollout.argoproj.io/revision"); revision != "" {
		return revision
	}
	
	return "0"
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := fmt.Sscanf(v, "%d", new(int)); err == nil && i == 1 {
				var result int
				fmt.Sscanf(v, "%d", &result)
				return result
			}
		}
	}
	return 0
}

func getRolloutImages(spec map[string]interface{}) map[string]string {
	images := make(map[string]string)
	
	if spec == nil {
		return images
	}
	
	// Navigate through spec.template.spec.containers to find images
	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		return images
	}
	
	templateSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return images
	}
	
	containers, ok := templateSpec["containers"].([]interface{})
	if !ok {
		return images
	}
	
	for _, containerInterface := range containers {
		container, ok := containerInterface.(map[string]interface{})
		if !ok {
			continue
		}
		
		name := getString(container, "name")
		image := getString(container, "image")
		
		if name != "" && image != "" {
			images[name] = image
		}
	}
	
	return images
}

// JWT client/server configuration structures
type JWTClientConfig struct {
	Platform      string `json:"platform"`
	Environment   string `json:"environment"`
	Team          string `json:"team"`
	Path          string `json:"path"`
	Owner         string `json:"owner"`
	LocalName     string `json:"localName"`
	TargetService string `json:"targetService"`
	Secret        string `json:"secret"`
}

type JWTServerConfig struct {
	Platform      string `json:"platform"`
	Environment   string `json:"environment"`
	Team          string `json:"team"`
	Path          string `json:"path"`
	Owner         string `json:"owner"`
	LocalName     string `json:"localName"`
	ServiceName   string `json:"serviceName"`
	ClientName    string `json:"clientName"`
	ClientSecret  string `json:"clientSecret"`
}

// CreateJWTClient creates a JWT client secret
func (a *App) CreateJWTClient(config JWTClientConfig) error {
	if config.Path == "" || config.Owner == "" || config.LocalName == "" || 
	   config.TargetService == "" || config.Secret == "" {
		return fmt.Errorf("all fields are required for JWT client creation")
	}

	// Build yak command
	args := []string{"secret", "jwt", "client", "--path", config.Path, "--owner", config.Owner,
		"--local-name", config.LocalName, "--target-service", config.TargetService, 
		"--secret", config.Secret}
	
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}

	// Execute yak secret jwt client
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create JWT client secret: %w", err)
	}

	return nil
}

// CreateJWTServer creates a JWT server secret
func (a *App) CreateJWTServer(config JWTServerConfig) error {
	if config.Path == "" || config.Owner == "" || config.LocalName == "" || 
	   config.ServiceName == "" || config.ClientName == "" || config.ClientSecret == "" {
		return fmt.Errorf("all fields are required for JWT server creation")
	}

	// Build yak command
	args := []string{"secret", "jwt", "server", "--path", config.Path, "--owner", config.Owner,
		"--local-name", config.LocalName, "--service-name", config.ServiceName, 
		"--client-name", config.ClientName, "--client-secret", config.ClientSecret}
	
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if config.Team != "" {
		args = append(args, "--team", config.Team)
	}

	// Execute yak secret jwt server
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create JWT server secret: %w", err)
	}

	return nil
}