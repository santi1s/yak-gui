package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

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