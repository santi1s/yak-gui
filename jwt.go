package main

import (
	"fmt"
	"os/exec"
)

// JWT client/server configuration structures
type JWTClientConfig struct {
	Platform      string `json:"platform"`
	Environment   string `json:"environment"`
	Path          string `json:"path"`
	Owner         string `json:"owner"`
	LocalName     string `json:"localName"`
	TargetService string `json:"targetService"`
	Secret        string `json:"secret"`
}

type JWTServerConfig struct {
	Platform      string `json:"platform"`
	Environment   string `json:"environment"`
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

	// Execute yak secret jwt server
	cmd := exec.Command(findYakExecutable(), args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create JWT server secret: %w", err)
	}

	return nil
}