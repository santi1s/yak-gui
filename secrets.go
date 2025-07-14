package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// SecretConfig represents secret management configuration
type SecretConfig struct {
	Platform    string `json:"platform"`
	Environment string `json:"environment"`
}

// YakSecretConfig represents the structure of secret.yml
type YakSecretConfig struct {
	Clusters  map[string]ClusterConfig  `yaml:"clusters"`
	Platforms map[string]PlatformConfig `yaml:"platforms"`
}

type ClusterConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type PlatformConfig struct {
	Clusters             []string                       `yaml:"clusters"`
	AwsProfile           string                         `yaml:"awsProfile"`
	AwsRegion            string                         `yaml:"awsRegion"`
	VaultRole            string                         `yaml:"vaultRole"`
	Environments         map[string]string              `yaml:"environments"`
	VaultParentNamespace string                         `yaml:"vaultParentNamespace"`
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

// GetSecrets lists secrets from a path using the yak CLI and fetches metadata for each secret
func (a *App) GetSecrets(config SecretConfig, path string) ([]SecretListItem, error) {
	// Build yak command for listing secret paths
	args := []string{"secret", "list", "--json"}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
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

	// Parse JSON output to get secret paths
	var secretMap map[string]interface{}
	if err := json.Unmarshal(output, &secretMap); err != nil {
		return nil, fmt.Errorf("failed to parse yak secret list output: %w", err)
	}
	
	// Extract secret paths from the list response
	var secretPaths []string
	if keysArray, ok := secretMap["keys"].([]interface{}); ok {
		for _, keyItem := range keysArray {
			if keyStr, ok := keyItem.(string); ok {
				secretPaths = append(secretPaths, keyStr)
			}
		}
	}

	// Now fetch metadata for each secret path
	var secrets []SecretListItem
	for _, secretPath := range secretPaths {
		// Check if this is a folder (ends with /)
		isFolder := strings.HasSuffix(secretPath, "/")
		
		if isFolder {
			// For folders, create entry without metadata
			secret := SecretListItem{
				Path:      secretPath,
				Version:   0,
				Owner:     "Folder",
				Usage:     "Directory", 
				Source:    "Folder",
				CreatedAt: "",
				UpdatedAt: "",
			}
			secrets = append(secrets, secret)
		} else {
			// For actual secrets, fetch metadata
			// Build the full path for metadata request
			fullPath := secretPath
			if path != "" && !strings.HasPrefix(secretPath, path) {
				// If we're in a subdirectory, construct the full path
				fullPath = strings.TrimSuffix(path, "/") + "/" + secretPath
			}
			
			secret, err := a.getSecretMetadata(config, fullPath)
			if err != nil {
				// If metadata fetch fails, create a basic entry with unknown metadata
				fmt.Printf("Warning: failed to get metadata for secret %s: %v\n", secretPath, err)
				secret = SecretListItem{
					Path:      secretPath,
					Version:   0,
					Owner:     "Unknown",
					Usage:     "Unknown", 
					Source:    "Unknown",
					CreatedAt: "",
					UpdatedAt: "",
				}
			} else {
				// Ensure the returned secret has the relative path, not the full path
				secret.Path = secretPath
			}
			secrets = append(secrets, secret)
		}
	}

	return secrets, nil
}

// getSecretMetadata fetches metadata for a specific secret using yak secret metadata get
func (a *App) getSecretMetadata(config SecretConfig, secretPath string) (SecretListItem, error) {
	// Build yak command for getting secret metadata
	args := []string{"secret", "metadata", "get", "--json", "--path", secretPath}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}

	// Execute yak secret metadata get with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	output, err := cmd.Output()
	if err != nil {
		return SecretListItem{}, fmt.Errorf("yak secret metadata get failed: %w", err)
	}

	// Parse JSON output
	var metadataMap map[string]interface{}
	if err := json.Unmarshal(output, &metadataMap); err != nil {
		return SecretListItem{}, fmt.Errorf("failed to parse secret metadata: %w", err)
	}

	// Extract metadata fields (Path will be set by caller)
	secret := SecretListItem{
		Path:      "", // Will be set by caller to avoid path duplication
		Version:   getLatestVersion(metadataMap),
		Owner:     getOwnerFromMetadata(metadataMap),
		Usage:     getUsageFromMetadata(metadataMap),
		Source:    getSourceFromMetadata(metadataMap),
		CreatedAt: getCreatedAtFromMetadata(metadataMap),
		UpdatedAt: getUpdatedAtFromMetadata(metadataMap),
	}

	return secret, nil
}

// Helper functions to extract metadata from vault response
func getLatestVersion(metadata map[string]interface{}) int {
	if currentVersion := getInt(metadata, "current_version"); currentVersion > 0 {
		return currentVersion
	}
	return 1 // Default to version 1 if not found
}

func getOwnerFromMetadata(metadata map[string]interface{}) string {
	// Check custom_metadata first
	if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
		if owner := getString(customMetadata, "owner"); owner != "" {
			return owner
		}
	}
	return "Unknown"
}

func getUsageFromMetadata(metadata map[string]interface{}) string {
	// Check custom_metadata first
	if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
		if usage := getString(customMetadata, "usage"); usage != "" {
			return usage
		}
	}
	return "Unknown"
}

func getSourceFromMetadata(metadata map[string]interface{}) string {
	// Check custom_metadata first
	if customMetadata, ok := metadata["custom_metadata"].(map[string]interface{}); ok {
		if source := getString(customMetadata, "source"); source != "" {
			return source
		}
	}
	return "Unknown"
}

func getCreatedAtFromMetadata(metadata map[string]interface{}) string {
	if createdTime := getString(metadata, "created_time"); createdTime != "" {
		return createdTime
	}
	return ""
}

func getUpdatedAtFromMetadata(metadata map[string]interface{}) string {
	if updatedTime := getString(metadata, "updated_time"); updatedTime != "" {
		return updatedTime
	}
	return ""
}

// GetSecretData retrieves secret data from a specific path
func (a *App) GetSecretData(config SecretConfig, path string, version int) (*SecretData, error) {
	if path == "" {
		return nil, fmt.Errorf("secret path is required")
	}

	fmt.Printf("DEBUG: GetSecretData called with path='%s', version=%d\n", path, version)

	// Build yak command
	args := []string{"secret", "get", "--json", "--path", path}
	if config.Platform != "" {
		args = append(args, "--platform", config.Platform)
	}
	if config.Environment != "" {
		args = append(args, "--environment", config.Environment)
	}
	if version > 0 {
		args = append(args, "--version", fmt.Sprintf("%d", version))
	}

	fmt.Printf("DEBUG: yak secret get command: %v\n", args)

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

	fmt.Printf("DEBUG: yak secret get raw output: %s\n", string(output))

	// Parse JSON output - yak secret get returns the vault response format
	var vaultResponse map[string]interface{}
	if err := json.Unmarshal(output, &vaultResponse); err != nil {
		return nil, fmt.Errorf("failed to parse secret data: %w", err)
	}

	// Extract data and metadata from vault response
	var secretData SecretData
	secretData.Path = path
	secretData.Version = version
	secretData.Data = make(map[string]string) // Initialize to prevent nil map
	
	// Initialize metadata with safe defaults
	secretData.Metadata = SecretMetadata{
		Owner:     "Unknown",
		Usage:     "Unknown",
		Source:    "Unknown",
		Version:   version,
		Destroyed: false,
		CreatedAt: "",
		UpdatedAt: "",
	}

	// The yak secret get command returns the vault response in format:
	// {"data": {...secret key-value pairs...}, "metadata": {...metadata including custom_metadata...}}
	
	// Extract secret data from root-level "data" section
	if dataSection, ok := vaultResponse["data"].(map[string]interface{}); ok {
		fmt.Printf("DEBUG: Found data section with keys: %v\n", getMapKeysFromInterface(dataSection))
		for key, value := range dataSection {
			if strValue, ok := value.(string); ok {
				secretData.Data[key] = strValue
			} else {
				// Handle non-string values by converting to string
				secretData.Data[key] = fmt.Sprintf("%v", value)
			}
		}
	}
	
	// Extract metadata from root-level "metadata" section
	if metadataSection, ok := vaultResponse["metadata"].(map[string]interface{}); ok {
		fmt.Printf("DEBUG: Found metadata section with keys: %v\n", getMapKeysFromInterface(metadataSection))
		
		// Safely extract metadata fields
		if version := getInt(metadataSection, "version"); version > 0 {
			secretData.Metadata.Version = version
		}
		if createdTime := getString(metadataSection, "created_time"); createdTime != "" {
			secretData.Metadata.CreatedAt = createdTime
		}
		if updatedTime := getString(metadataSection, "updated_time"); updatedTime != "" {
			secretData.Metadata.UpdatedAt = updatedTime
		}
		secretData.Metadata.Destroyed = getBool(metadataSection, "destroyed")
		
		// Extract custom metadata (owner, usage, source) if available
		if customMetadata, ok := metadataSection["custom_metadata"].(map[string]interface{}); ok {
			fmt.Printf("DEBUG: Found custom_metadata with keys: %v\n", getMapKeysFromInterface(customMetadata))
			fmt.Printf("DEBUG: custom_metadata content: %+v\n", customMetadata)
			if owner := getString(customMetadata, "owner"); owner != "" {
				fmt.Printf("DEBUG: Extracted owner: %s\n", owner)
				secretData.Metadata.Owner = owner
			} else {
				fmt.Printf("DEBUG: Owner not found or empty in custom_metadata\n")
			}
			if usage := getString(customMetadata, "usage"); usage != "" {
				fmt.Printf("DEBUG: Extracted usage: %s\n", usage)
				secretData.Metadata.Usage = usage
			} else {
				fmt.Printf("DEBUG: Usage not found or empty in custom_metadata\n")
			}
			if source := getString(customMetadata, "source"); source != "" {
				fmt.Printf("DEBUG: Extracted source: %s\n", source)
				secretData.Metadata.Source = source
			} else {
				fmt.Printf("DEBUG: Source not found or empty in custom_metadata\n")
			}
		} else {
			fmt.Printf("DEBUG: No custom_metadata found in metadata section\n")
			fmt.Printf("DEBUG: metadata section keys: %v\n", getMapKeysFromInterface(metadataSection))
			fmt.Printf("DEBUG: metadata section content: %+v\n", metadataSection)
		}
	} else {
		fmt.Printf("DEBUG: No metadata section found at root level\n")
		fmt.Printf("DEBUG: Root response keys: %v\n", getMapKeys(vaultResponse))
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

// LoadSecretConfig loads the secret.yml configuration file
func (a *App) LoadSecretConfig() (*YakSecretConfig, error) {
	// Try to find the secret.yml file
	var configPath string
	
	// First try TFINFRA_REPOSITORY_PATH/setup/yak_config/secret.yml
	if tfinfraPath := os.Getenv("TFINFRA_REPOSITORY_PATH"); tfinfraPath != "" {
		configPath = filepath.Join(tfinfraPath, "setup", "yak_config", "secret.yml")
		if _, err := os.Stat(configPath); err == nil {
			goto LoadConfig
		}
	}
	
	// Then try ~/.yak/secret.yml
	if homeDir, err := os.UserHomeDir(); err == nil {
		configPath = filepath.Join(homeDir, ".yak", "secret.yml")
		if _, err := os.Stat(configPath); err == nil {
			goto LoadConfig
		}
	}
	
	// Finally try ./secret.yml
	configPath = "secret.yml"
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("secret.yml not found in any expected location")
	}
	
LoadConfig:
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret config file: %w", err)
	}
	
	// Parse the YAML
	var config YakSecretConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse secret config file: %w", err)
	}
	
	return &config, nil
}

// GetSecretConfigPlatforms returns the list of available platforms
func (a *App) GetSecretConfigPlatforms() ([]string, error) {
	config, err := a.LoadSecretConfig()
	if err != nil {
		return nil, err
	}
	
	platforms := make([]string, 0, len(config.Platforms))
	for platform := range config.Platforms {
		platforms = append(platforms, platform)
	}
	
	sort.Strings(platforms)
	return platforms, nil
}

// GetSecretConfigEnvironments returns the list of available environments for a given platform
func (a *App) GetSecretConfigEnvironments(platform string) ([]string, error) {
	config, err := a.LoadSecretConfig()
	if err != nil {
		return nil, err
	}
	
	platformConfig, exists := config.Platforms[platform]
	if !exists {
		return nil, fmt.Errorf("platform %s not found", platform)
	}
	
	environments := make([]string, 0, len(platformConfig.Environments))
	for env := range platformConfig.Environments {
		environments = append(environments, env)
	}
	
	sort.Strings(environments)
	return environments, nil
}

// GetSecretConfigPaths returns available path prefixes by running yak secret list
func (a *App) GetSecretConfigPaths(platform, environment string) ([]string, error) {
	if platform == "" {
		return []string{""}, nil // Return empty path if no platform specified
	}
	
	// Build yak secret list command
	args := []string{"secret", "list", "--platform", platform}
	if environment != "" {
		args = append(args, "--environment", environment)
	}
	args = append(args, "--json")
	
	fmt.Printf("DEBUG: yak secret list command: %v\n", args)
	
	// Execute yak secret list with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	output, err := cmd.Output()
	if err != nil {
		// If command fails, return empty path as fallback
		fmt.Printf("DEBUG: yak secret list failed: %v\n", err)
		return []string{""}, nil
	}
	
	fmt.Printf("DEBUG: yak secret list raw output: %s\n", string(output))
	
	// Parse JSON output - yak secret list returns {keys: [...]}
	var secretsResponse struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(output, &secretsResponse); err != nil {
		// If parsing fails, return empty path as fallback
		fmt.Printf("DEBUG: failed to parse secret list output: %v\n", err)
		return []string{""}, nil
	}
	
	fmt.Printf("DEBUG: parsed %d keys from list\n", len(secretsResponse.Keys))
	
	// Extract paths that end with "/"
	pathsSet := make(map[string]bool)
	pathsSet[""] = true // Always include empty path for "all paths"
	
	for _, key := range secretsResponse.Keys {
		if strings.HasSuffix(key, "/") {
			fmt.Printf("DEBUG: found directory path: %s\n", key)
			pathsSet[key] = true
		}
	}
	
	// Convert to sorted slice
	paths := make([]string, 0, len(pathsSet))
	for path := range pathsSet {
		paths = append(paths, path)
	}
	
	sort.Strings(paths)
	fmt.Printf("DEBUG: final paths: %+v\n", paths)
	return paths, nil
}

// getMapKeysFromInterface returns the keys of a map[string]interface{} for debugging
func getMapKeysFromInterface(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getMapKeys returns the keys of a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}