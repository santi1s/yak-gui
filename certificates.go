package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Certificate represents a certificate configuration
type Certificate struct {
	Name        string            `json:"name"`
	Conf        string            `json:"conf"`
	Issuer      string            `json:"issuer"`
	Tags        []string          `json:"tags"`
	Cloudflare  CloudflareConfig  `json:"cloudflare"`
	Secret      SecretPath        `json:"secret"`
}

// CloudflareConfig represents cloudflare configuration for a certificate
type CloudflareConfig struct {
	Path string `json:"path"`
	Zone string `json:"zone"`
}

// SecretPath represents the secret storage configuration
type SecretPath struct {
	Platform string            `json:"platform"`
	Env      string            `json:"env"`
	Path     string            `json:"path"`
	Keys     map[string]string `json:"keys"`
}

// CertificateStatus represents the status of a certificate
type CertificateStatus struct {
	Name         string    `json:"name"`
	Issuer       string    `json:"issuer"`
	Subject      string    `json:"subject"`
	Expiration   time.Time `json:"expiration"`
	DaysUntilExp int       `json:"daysUntilExpiration"`
	Status       string    `json:"status"`
	SerialNumber string    `json:"serialNumber"`
}

// CertificateOperation represents an operation result
type CertificateOperation struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

// CheckGandiToken verifies the GANDI_TOKEN is configured correctly
func (a *App) CheckGandiToken() (*CertificateOperation, error) {
	// Execute yak certificate gandi-check
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), "certificate", "gandi-check")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &CertificateOperation{
			Success: false,
			Message: "Failed to check Gandi token",
			Output:  string(output),
		}, nil
	}

	return &CertificateOperation{
		Success: true,
		Message: "Gandi token is valid",
		Output:  string(output),
	}, nil
}

// GetCertificateConfig retrieves the certificate configuration from terraform-infra
func (a *App) GetCertificateConfig() ([]Certificate, error) {
	// This would typically read from the config.yml file
	// For now, we'll return an empty list and let users manually configure
	// In a real implementation, you might want to:
	// 1. Clone/pull terraform-infra repo
	// 2. Parse the config.yml file
	// 3. Return the parsed certificates
	
	return []Certificate{}, nil
}

// RenewCertificate initiates the certificate renewal process
func (a *App) RenewCertificate(certificateName, jiraTicket string) (*CertificateOperation, error) {
	if certificateName == "" {
		return nil, fmt.Errorf("certificate name is required")
	}
	if jiraTicket == "" {
		return nil, fmt.Errorf("JIRA ticket is required")
	}

	// Build yak command
	args := []string{"certificate", "renew", "--certificate", certificateName, "-j", jiraTicket}
	
	// Execute yak certificate renew with timeout (this can take a while)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &CertificateOperation{
			Success: false,
			Message: fmt.Sprintf("Failed to renew certificate %s", certificateName),
			Output:  string(output),
		}, nil
	}

	return &CertificateOperation{
		Success: true,
		Message: fmt.Sprintf("Certificate %s renewal initiated successfully", certificateName),
		Output:  string(output),
	}, nil
}

// RefreshCertificateSecret refreshes the secret with the new certificate
func (a *App) RefreshCertificateSecret(certificateName, jiraTicket string) (*CertificateOperation, error) {
	if certificateName == "" {
		return nil, fmt.Errorf("certificate name is required")
	}
	if jiraTicket == "" {
		return nil, fmt.Errorf("JIRA ticket is required")
	}

	// Build yak command
	args := []string{"certificate", "refresh-secret", "--certificate", certificateName, "-j", jiraTicket}
	
	// Execute yak certificate refresh-secret with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &CertificateOperation{
			Success: false,
			Message: fmt.Sprintf("Failed to refresh secret for certificate %s", certificateName),
			Output:  string(output),
		}, nil
	}

	return &CertificateOperation{
		Success: true,
		Message: fmt.Sprintf("Secret for certificate %s refreshed successfully", certificateName),
		Output:  string(output),
	}, nil
}

// DescribeCertificateSecret describes the certificate secret details
func (a *App) DescribeCertificateSecret(certificateName string, version int, diffVersion int) (*CertificateOperation, error) {
	if certificateName == "" {
		return nil, fmt.Errorf("certificate name is required")
	}

	// Build yak command
	args := []string{"certificate", "describe-secret", "-C", certificateName}
	
	if version > 0 {
		args = append(args, "-v", fmt.Sprintf("%d", version))
	}
	
	if diffVersion > 0 {
		args = append(args, fmt.Sprintf("--diff-version=%d", diffVersion))
	}
	
	// Execute yak certificate describe-secret
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, findYakExecutable(), args...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &CertificateOperation{
			Success: false,
			Message: fmt.Sprintf("Failed to describe secret for certificate %s", certificateName),
			Output:  string(output),
		}, nil
	}

	return &CertificateOperation{
		Success: true,
		Message: fmt.Sprintf("Certificate %s secret details retrieved", certificateName),
		Output:  string(output),
	}, nil
}

// ListCertificates lists available certificates (this would typically parse the config)
func (a *App) ListCertificates() ([]string, error) {
	// For now, return some example certificate names
	// In a real implementation, this would parse the config.yml file
	examples := []string{
		"keyless-staging-doctolib.de",
		"keyless-prod-doctolib.fr", 
		"wildcard-doctolib.com",
		"api-doctolib.net",
	}
	
	return examples, nil
}

// SendCertificateNotification sends email notification to technical services
func (a *App) SendCertificateNotification(certificateName, operationDate, operation string) (*CertificateOperation, error) {
	// This is a placeholder - in a real implementation you might:
	// 1. Use a mail service/API
	// 2. Generate an email template
	// 3. Send to technicalservices-all@doctolib.com
	
	message := fmt.Sprintf(`
Email would be sent to: technicalservices-all@doctolib.com

Subject: SSL Certificate %s - %s Scheduled for %s

Dear Technical Services Team,

This is to inform you that we will be performing an SSL certificate %s operation:

Certificate: %s
Operation: %s
Scheduled Date: %s

Please be aware that some customers may need to manually download our certificate and upload it to their trust store.

Best regards
`, operation, certificateName, operationDate, operation, certificateName, operation, operationDate)

	return &CertificateOperation{
		Success: true,
		Message: "Email notification template generated successfully",
		Output:  message,
	}, nil
}