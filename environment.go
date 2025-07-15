package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EnvironmentProfile represents a saved environment configuration
type EnvironmentProfile struct {
	Name                    string `json:"name"`
	AWSProfile              string `json:"aws_profile"`
	Kubeconfig              string `json:"kubeconfig"`
	PATH                    string `json:"path"`
	TfInfraRepositoryPath   string `json:"tf_infra_repository_path"`
	GandiToken              string `json:"gandi_token"`
	CreatedAt               string `json:"created_at"`
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

// SetGandiToken sets the GANDI_TOKEN environment variable for the current session
func (a *App) SetGandiToken(token string) error {
	if token == "" {
		return fmt.Errorf("GANDI_TOKEN cannot be empty")
	}
	return os.Setenv("GANDI_TOKEN", token)
}

// GetGandiToken returns the current GANDI_TOKEN environment variable
func (a *App) GetGandiToken() string {
	return os.Getenv("GANDI_TOKEN")
}

// IsGandiTokenSet returns whether GANDI_TOKEN is set without revealing the value
func (a *App) IsGandiTokenSet() bool {
	token := os.Getenv("GANDI_TOKEN")
	return token != ""
}

// GetAWSProfiles reads ~/.aws/config and returns available profiles (excluding -sso profiles)
func (a *App) GetAWSProfiles() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}
	
	configPath := filepath.Join(homeDir, ".aws", "config")
	
	// Debug logging to help troubleshoot Finder launch issues
	fmt.Printf("DEBUG: Looking for AWS config at: %s\n", configPath)
	fmt.Printf("DEBUG: Home directory: %s\n", homeDir)
	
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("DEBUG: AWS config file not found at %s\n", configPath)
			return []string{}, nil // Return empty list if config doesn't exist
		}
		return nil, fmt.Errorf("failed to open AWS config file at %s: %v", configPath, err)
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

// GetShellEnvironment executes shell commands to get the full environment
func (a *App) GetShellEnvironment() (map[string]string, error) {
	// Determine user's shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh" // Default to zsh on macOS
	}
	
	// Get home directory for debugging
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	
	// Try multiple approaches to get full environment
	var output []byte
	var cmdErr error
	
	// Approach 1: Interactive login shell (most comprehensive)
	fmt.Printf("DEBUG: Trying interactive login shell: %s -l -i -c env\n", shell)
	cmd := exec.Command(shell, "-l", "-i", "-c", "env")
	cmd.Dir = homeDir // Set working directory to home
	output, cmdErr = cmd.Output()
	
	if cmdErr != nil {
		// Approach 2: Login shell
		fmt.Printf("DEBUG: Interactive failed, trying login shell: %s -l -c env\n", shell)
		cmd = exec.Command(shell, "-l", "-c", "env")
		cmd.Dir = homeDir
		output, cmdErr = cmd.Output()
	}
	
	if cmdErr != nil {
		// Approach 3: Source .zshrc explicitly
		fmt.Printf("DEBUG: Login failed, trying explicit zshrc sourcing\n")
		cmd = exec.Command(shell, "-c", "source ~/.zshrc && env")
		cmd.Dir = homeDir
		output, cmdErr = cmd.Output()
	}
	
	if cmdErr != nil {
		return nil, fmt.Errorf("failed to get shell environment with all methods: %v", cmdErr)
	}
	
	// Parse environment variables
	envVars := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}
	
	// Debug: Print some key variables we found
	fmt.Printf("DEBUG: Found %d environment variables\n", len(envVars))
	for _, key := range []string{"PATH", "AWS_PROFILE", "TFINFRA_REPOSITORY_PATH", "HOME", "GANDI_TOKEN"} {
		if value, exists := envVars[key]; exists {
			// Don't log sensitive tokens in full, just first/last chars
			if key == "GANDI_TOKEN" && len(value) > 8 {
				fmt.Printf("DEBUG: %s=%s...%s\n", key, value[:4], value[len(value)-4:])
			} else {
				fmt.Printf("DEBUG: %s=%s\n", key, value)
			}
		} else {
			fmt.Printf("DEBUG: %s not found in shell environment\n", key)
		}
	}
	
	return envVars, nil
}

// ImportShellEnvironment imports environment variables from shell and sets them in the current process
func (a *App) ImportShellEnvironment() error {
	shellEnv, err := a.GetShellEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get shell environment: %v", err)
	}
	
	// Import key environment variables
	importantVars := []string{
		"AWS_PROFILE",
		"KUBECONFIG", 
		"PATH",
		"TFINFRA_REPOSITORY_PATH",
		"HOME",
		"GANDI_TOKEN",
	}
	
	for _, varName := range importantVars {
		if value, exists := shellEnv[varName]; exists && value != "" {
			if err := os.Setenv(varName, value); err != nil {
				fmt.Printf("Warning: failed to set %s: %v\n", varName, err)
			} else {
				fmt.Printf("Imported %s=%s\n", varName, value)
			}
		}
	}
	
	return nil
}

// GetEnvironmentVariables returns a map of current environment variables (with sensitive values masked)
func (a *App) GetEnvironmentVariables() map[string]string {
	envVars := map[string]string{
		"AWS_PROFILE":              os.Getenv("AWS_PROFILE"),
		"KUBECONFIG":               os.Getenv("KUBECONFIG"),
		"HOME":                     os.Getenv("HOME"),
		"PATH":                     os.Getenv("PATH"),
		"TFINFRA_REPOSITORY_PATH":  os.Getenv("TFINFRA_REPOSITORY_PATH"),
		"GANDI_TOKEN":              maskSensitiveValue(os.Getenv("GANDI_TOKEN")),
	}
	
	// Debug logging to help troubleshoot Finder launch issues
	fmt.Printf("DEBUG: Environment variables when called:\n")
	for key, value := range envVars {
		fmt.Printf("  %s=%s\n", key, value)
	}
	
	return envVars
}

// maskSensitiveValue masks sensitive values by showing only first 4 and last 4 characters
func maskSensitiveValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****" // Mask entirely if too short
	}
	return value[:4] + "..." + value[len(value)-4:]
}

// SaveEnvironmentProfile saves the current environment configuration as a profile
func (a *App) SaveEnvironmentProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	
	// Get current environment variables (DO NOT store sensitive tokens in profiles)
	profile := EnvironmentProfile{
		Name:                  name,
		AWSProfile:            os.Getenv("AWS_PROFILE"),
		Kubeconfig:            os.Getenv("KUBECONFIG"),
		PATH:                  os.Getenv("PATH"),
		TfInfraRepositoryPath: os.Getenv("TFINFRA_REPOSITORY_PATH"),
		GandiToken:            "", // Never save sensitive tokens in profiles
		CreatedAt:             time.Now().Format(time.RFC3339),
	}
	
	// Get home directory for storing profiles
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	
	// Create .yak-gui directory if it doesn't exist
	yakGuiDir := filepath.Join(homeDir, ".yak-gui")
	if err := os.MkdirAll(yakGuiDir, 0755); err != nil {
		return fmt.Errorf("failed to create .yak-gui directory: %v", err)
	}
	
	// Load existing profiles
	profiles, err := a.GetEnvironmentProfiles()
	if err != nil {
		profiles = []EnvironmentProfile{}
	}
	
	// Update or add profile
	found := false
	for i, p := range profiles {
		if p.Name == name {
			profiles[i] = profile
			found = true
			break
		}
	}
	if !found {
		profiles = append(profiles, profile)
	}
	
	// Save profiles to JSON file
	profilesPath := filepath.Join(yakGuiDir, "environment-profiles.json")
	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %v", err)
	}
	
	if err := os.WriteFile(profilesPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profiles file: %v", err)
	}
	
	return nil
}

// GetEnvironmentProfiles returns all saved environment profiles
func (a *App) GetEnvironmentProfiles() ([]EnvironmentProfile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	
	profilesPath := filepath.Join(homeDir, ".yak-gui", "environment-profiles.json")
	
	// Check if file exists
	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		return []EnvironmentProfile{}, nil
	}
	
	// Read profiles file
	data, err := os.ReadFile(profilesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles file: %v", err)
	}
	
	var profiles []EnvironmentProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profiles: %v", err)
	}
	
	return profiles, nil
}

// LoadEnvironmentProfile loads a saved environment profile and applies it
func (a *App) LoadEnvironmentProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	
	profiles, err := a.GetEnvironmentProfiles()
	if err != nil {
		return fmt.Errorf("failed to get profiles: %v", err)
	}
	
	// Find the profile
	var targetProfile *EnvironmentProfile
	for _, p := range profiles {
		if p.Name == name {
			targetProfile = &p
			break
		}
	}
	
	if targetProfile == nil {
		return fmt.Errorf("profile '%s' not found", name)
	}
	
	// Apply the profile environment variables (skip sensitive tokens)
	envVars := map[string]string{
		"AWS_PROFILE":            targetProfile.AWSProfile,
		"KUBECONFIG":             targetProfile.Kubeconfig,
		"PATH":                   targetProfile.PATH,
		"TFINFRA_REPOSITORY_PATH": targetProfile.TfInfraRepositoryPath,
		// Note: GANDI_TOKEN is not restored from profiles for security
	}
	
	for key, value := range envVars {
		if value != "" {
			if err := os.Setenv(key, value); err != nil {
				fmt.Printf("Warning: failed to set %s: %v\n", key, err)
			}
		}
	}
	
	return nil
}

// DeleteEnvironmentProfile deletes a saved environment profile
func (a *App) DeleteEnvironmentProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	
	profiles, err := a.GetEnvironmentProfiles()
	if err != nil {
		return fmt.Errorf("failed to get profiles: %v", err)
	}
	
	// Remove the profile
	var updatedProfiles []EnvironmentProfile
	found := false
	for _, p := range profiles {
		if p.Name != name {
			updatedProfiles = append(updatedProfiles, p)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("profile '%s' not found", name)
	}
	
	// Save updated profiles
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	
	profilesPath := filepath.Join(homeDir, ".yak-gui", "environment-profiles.json")
	data, err := json.MarshalIndent(updatedProfiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %v", err)
	}
	
	if err := os.WriteFile(profilesPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profiles file: %v", err)
	}
	
	return nil
}