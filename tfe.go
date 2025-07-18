package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TFEConfig represents TFE connection configuration
type TFEConfig struct {
	Endpoint     string `json:"endpoint"`
	Organization string `json:"organization"`
	Token        string `json:"token,omitempty"`
}

// TFEWorkspace represents a TFE workspace
type TFEWorkspace struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Environment      string            `json:"environment,omitempty"`
	TerraformVersion string            `json:"terraformVersion,omitempty"`
	Status           string            `json:"status"` // active, locked, disabled
	LastRun          string            `json:"lastRun,omitempty"`
	Owner            string            `json:"owner,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Organization     string            `json:"organization"`
	CreatedAt        string            `json:"createdAt,omitempty"`
	UpdatedAt        string            `json:"updatedAt,omitempty"`
	AutoApply        bool              `json:"autoApply"`
	TerraformWorking bool              `json:"terraformWorking"`
	VCSRepo          *TFEVCSRepo       `json:"vcsRepo,omitempty"`
	Variables        map[string]string `json:"variables,omitempty"`
}

// TFEVCSRepo represents VCS repository information
type TFEVCSRepo struct {
	Identifier        string `json:"identifier"`
	Branch            string `json:"branch"`
	IngressSubmodules bool   `json:"ingressSubmodules"`
}

// TFERun represents a TFE run
type TFERun struct {
	ID               string `json:"id"`
	WorkspaceID      string `json:"workspaceId"`
	WorkspaceName    string `json:"workspaceName"`
	Status           string `json:"status"` // pending, planning, planned, applying, applied, discarded, errored, canceled
	CreatedAt        string `json:"createdAt"`
	Message          string `json:"message,omitempty"`
	Source           string `json:"source"` // manual, vcs, api
	TerraformVersion string `json:"terraformVersion,omitempty"`
	HasChanges       bool   `json:"hasChanges"`
	IsDestroy        bool   `json:"isDestroy"`
	IsConfirmable    bool   `json:"isConfirmable"`
	Actions          struct {
		IsConfirmable bool `json:"isConfirmable"`
		IsCancelable  bool `json:"isCancelable"`
		IsDiscardable bool `json:"isDiscardable"`
	} `json:"actions"`
	CreatedBy string `json:"createdBy,omitempty"`
	URL       string `json:"url,omitempty"`
}

// TFEPlanExecution represents a plan execution request
type TFEPlanExecution struct {
	WorkspaceNames   []string `json:"workspaceNames,omitempty"`
	Owner            string   `json:"owner,omitempty"`
	TerraformVersion string   `json:"terraformVersion"`
	Message          string   `json:"message,omitempty"`
	Wait             bool     `json:"wait"`
}

// TFEPlanResult represents the result of a plan execution
type TFEPlanResult struct {
	WorkspaceName string    `json:"workspaceName"`
	RunID         string    `json:"runId"`
	Status        string    `json:"status"`
	HasChanges    bool      `json:"hasChanges"`
	Message       string    `json:"message,omitempty"`
	Error         string    `json:"error,omitempty"`
	URL           string    `json:"url,omitempty"`
	Duration      string    `json:"duration,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// TFEVersionInfo represents Terraform version information
type TFEVersionInfo struct {
	Version    string `json:"version"`
	Status     string `json:"status"` // enabled, disabled, deprecated
	IsDefault  bool   `json:"isDefault"`
	IsSupported bool   `json:"isSupported"`
	Beta       bool   `json:"beta"`
	Usage      int    `json:"usage"` // Number of workspaces using this version
}

// GetTFEWorkspaces retrieves all TFE workspaces
func (a *App) GetTFEWorkspaces(config TFEConfig) ([]TFEWorkspace, error) {
	// Build yak command
	args := []string{"tfe", "workspace", "list", "--json"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list TFE workspaces: %w - %s", err, string(output))
	}
	
	// Parse the JSON output which is an array of workspace names
	var workspaceNames []string
	if err := json.Unmarshal(output, &workspaceNames); err != nil {
		return nil, fmt.Errorf("failed to parse TFE workspaces response: %w", err)
	}
	
	// Convert workspace names to TFEWorkspace objects
	var workspaces []TFEWorkspace
	for _, name := range workspaceNames {
		workspace := TFEWorkspace{
			ID:           name, // Use name as ID for now
			Name:         name,
			Organization: config.Organization,
			Status:       "active", // Default status
			// Don't extract fake tags - TFE has real tags but we can't get them from the list command
			Environment: extractEnvironmentFromName(name),
			Owner:       extractOwnerFromName(name),
			Tags:        []string{}, // Empty tags since we can't get real ones from the list command
		}
		workspaces = append(workspaces, workspace)
	}
	
	return workspaces, nil
}

// Helper function to extract environment from workspace name
func extractEnvironmentFromName(name string) string {
	// Check for regional patterns first (more specific)
	if strings.Contains(name, "dev-aws-fr-par-1") {
		return "dev-aws-fr-par-1"
	} else if strings.Contains(name, "dev-aws-de-fra-1") {
		return "dev-aws-de-fra-1"
	} else if strings.Contains(name, "dev-aws-global") {
		return "dev-aws-global"
	} else if strings.Contains(name, "staging-aws-fr-par-1") {
		return "staging-aws-fr-par-1"
	} else if strings.Contains(name, "staging-aws-de-fra-1") {
		return "staging-aws-de-fra-1"
	} else if strings.Contains(name, "staging-aws-global") {
		return "staging-aws-global"
	} else if strings.Contains(name, "prod-aws-fr-par-1") || strings.Contains(name, "prd-aws-fr-par-1") {
		return "prod-aws-fr-par-1"
	} else if strings.Contains(name, "prod-aws-de-fra-1") || strings.Contains(name, "prd-aws-de-fra-1") {
		return "prod-aws-de-fra-1"
	} else if strings.Contains(name, "prod-aws-global") || strings.Contains(name, "prd-aws-global") {
		return "prod-aws-global"
	} else if strings.Contains(name, "preprod-aws-fr-par-1") {
		return "preprod-aws-fr-par-1"
	} else if strings.Contains(name, "preprod-aws-de-fra-1") {
		return "preprod-aws-de-fra-1"
	} else if strings.Contains(name, "preprod-aws-global") {
		return "preprod-aws-global"
	}
	
	// Check for general environment patterns
	if strings.Contains(name, "preprod") {
		return "preprod"
	} else if strings.Contains(name, "shared") {
		return "shared"
	} else if strings.Contains(name, "prd") || strings.Contains(name, "prod") {
		return "production"
	} else if strings.Contains(name, "staging") {
		return "staging"
	} else if strings.Contains(name, "dev") {
		return "development"
	} else if strings.Contains(name, "test") {
		return "testing"
	}
	
	return "unknown"
}

// Helper function to extract owner from workspace name
func extractOwnerFromName(name string) string {
	// Try to extract team/owner from workspace name patterns
	if strings.Contains(name, "tooling") {
		return "tooling-team"
	} else if strings.Contains(name, "security") {
		return "security-team"
	} else if strings.Contains(name, "logging") {
		return "logging-team"
	} else if strings.Contains(name, "cicd") {
		return "cicd-team"
	}
	return "unknown"
}


// GetTFEWorkspacesByTag retrieves TFE workspaces filtered by tag
func (a *App) GetTFEWorkspacesByTag(config TFEConfig, tag string, not bool) ([]TFEWorkspace, error) {
	// Build yak command
	args := []string{"tfe", "workspace", "list", "--json"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add tag filter
	if tag != "" {
		args = append(args, "--tag", tag)
		if not {
			args = append(args, "--not")
		}
	}
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list TFE workspaces by tag: %w - %s", err, string(output))
	}
	
	var workspaces []TFEWorkspace
	if err := json.Unmarshal(output, &workspaces); err != nil {
		return nil, fmt.Errorf("failed to parse TFE workspaces response: %w", err)
	}
	
	return workspaces, nil
}

// ExecuteTFEPlan executes a plan on TFE workspaces
func (a *App) ExecuteTFEPlan(config TFEConfig, execution TFEPlanExecution) ([]TFEPlanResult, error) {
	// Build yak command
	args := []string{"tfe", "plan"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add terraform version (required)
	args = append(args, "--version", execution.TerraformVersion)
	
	// Add workspaces or owner
	if len(execution.WorkspaceNames) > 0 {
		args = append(args, "--workspaces", strings.Join(execution.WorkspaceNames, ","))
	} else if execution.Owner != "" {
		args = append(args, "--owner", execution.Owner)
	}
	
	// Add wait flag if specified
	if execution.Wait {
		args = append(args, "--wait")
	}
	
	// Add JSON output
	args = append(args, "--json")
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes timeout for plan execution
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute TFE plan: %w - %s", err, string(output))
	}
	
	var results []TFEPlanResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse TFE plan response: %w", err)
	}
	
	return results, nil
}

// GetTFERuns retrieves TFE runs (currently not supported by yak CLI)
func (a *App) GetTFERuns(config TFEConfig, workspaceID string) ([]TFERun, error) {
	// Note: The yak tfe run command doesn't have a list subcommand
	// For now, return empty array until the yak CLI supports this
	return []TFERun{}, nil
}

// LockTFEWorkspace locks a TFE workspace
func (a *App) LockTFEWorkspace(config TFEConfig, workspaceNames []string, checkStatus bool) error {
	// Build yak command
	args := []string{"tfe", "workspace", "lock"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add workspaces
	if len(workspaceNames) > 0 {
		args = append(args, "--workspaces", strings.Join(workspaceNames, ","))
	}
	
	// Add check status flag
	if checkStatus {
		args = append(args, "--check-status")
	}
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to lock TFE workspace: %w - %s", err, string(output))
	}
	
	return nil
}

// UnlockTFEWorkspace unlocks a TFE workspace
func (a *App) UnlockTFEWorkspace(config TFEConfig, workspaceNames []string, force bool) error {
	// Build yak command
	args := []string{"tfe", "workspace", "unlock"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add workspaces
	if len(workspaceNames) > 0 {
		args = append(args, "--workspaces", strings.Join(workspaceNames, ","))
	}
	
	// Add force flag
	if force {
		args = append(args, "--force")
	}
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unlock TFE workspace: %w - %s", err, string(output))
	}
	
	return nil
}

// SetTFEWorkspaceVersion sets the Terraform version for TFE workspaces
func (a *App) SetTFEWorkspaceVersion(config TFEConfig, workspaceNames []string, version string) error {
	// Build yak command
	args := []string{"tfe", "workspace", "set-version"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add workspaces
	if len(workspaceNames) > 0 {
		args = append(args, "--workspaces", strings.Join(workspaceNames, ","))
	}
	
	// Add version
	args = append(args, "--version", version)
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set TFE workspace version: %w - %s", err, string(output))
	}
	
	return nil
}

// DiscardTFERuns discards old TFE runs
func (a *App) DiscardTFERuns(config TFEConfig, ageHours int, discardPending bool, dryRun bool, allWorkspaces bool) error {
	// Build yak command
	args := []string{"tfe", "run", "discard"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add age threshold
	args = append(args, "--age", fmt.Sprintf("%d", ageHours))
	
	// Add flags
	if discardPending {
		args = append(args, "--discard-pending")
	}
	
	if dryRun {
		args = append(args, "--dry-run")
	}
	
	if allWorkspaces {
		args = append(args, "--all-workspaces")
	}
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes timeout
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to discard TFE runs: %w - %s", err, string(output))
	}
	
	return nil
}

// GetTFEVersions retrieves available Terraform versions
func (a *App) GetTFEVersions(config TFEConfig) ([]TFEVersionInfo, error) {
	// Build yak command
	args := []string{"tfe", "versions", "list", "--json"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list TFE versions: %w - %s", err, string(output))
	}
	
	var versions []TFEVersionInfo
	if err := json.Unmarshal(output, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse TFE versions response: %w", err)
	}
	
	return versions, nil
}

// CheckTFEDeprecatedVersions checks for workspaces using deprecated Terraform versions
func (a *App) CheckTFEDeprecatedVersions(config TFEConfig, versionFile string, teamsFile string, sendEmail bool) (map[string]interface{}, error) {
	// Build yak command
	args := []string{"tfe", "check-versions"}
	
	// Add organization if specified
	if config.Organization != "" {
		args = append(args, "--organization", config.Organization)
	}
	
	// Add required files
	args = append(args, "--file", versionFile)
	args = append(args, "--teams", teamsFile)
	
	// Add send email flag
	if sendEmail {
		args = append(args, "--send-email")
	}
	
	// Add JSON output
	args = append(args, "--json")
	
	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	// Set environment variables for TFE authentication and ensure proper environment
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("TFE_ENDPOINT=%s", config.Endpoint),
		fmt.Sprintf("TFE_TOKEN=%s", config.Token),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to check TFE deprecated versions: %w - %s", err, string(output))
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TFE deprecated versions response: %w", err)
	}
	
	return result, nil
}

// GetTFEConfig retrieves TFE configuration from environment variables
func (a *App) GetTFEConfig() (TFEConfig, error) {
	config := TFEConfig{
		Endpoint:     "tfe.doctolib.net", // Default TFE endpoint for Doctolib
		Organization: "doctolib",         // Default organization
	}
	
	// Get environment variables
	env := a.GetEnvironmentVariables()
	
	// Override with environment variables if present
	if endpoint := env["TFE_ENDPOINT"]; endpoint != "" {
		config.Endpoint = endpoint
	}
	
	if org := env["TFE_ORGANIZATION"]; org != "" {
		config.Organization = org
	}
	
	if token := env["TFE_TOKEN"]; token != "" {
		config.Token = token
	}
	
	// Also check for TF_TOKEN_<hostname> format
	if config.Token == "" {
		// Extract hostname from endpoint for token lookup
		hostname := config.Endpoint
		if hostname != "" {
			// Try TF_TOKEN_<hostname> format (replace dots with underscores)
			tokenKey := "TF_TOKEN_" + strings.ReplaceAll(hostname, ".", "_")
			if token := env[tokenKey]; token != "" {
				config.Token = token
			}
		}
	}
	
	return config, nil
}

// SetTFEConfig sets TFE configuration in environment variables
func (a *App) SetTFEConfig(config TFEConfig) error {
	// Note: In a real application, you might want to store this more securely
	// For now, we'll just return nil as the config is passed from the frontend
	return nil
}