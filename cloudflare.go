package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

type CloudflareAPIConfig struct {
	Token       string `json:"token"`
	AccountName string `json:"accountName"`
}

type CloudflareOwner struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

type CloudflarePlan struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Currency         string `json:"currency"`
	LegacyID         string `json:"legacy_id"`
	IsSubscribed     bool   `json:"is_subscribed"`
	CanSubscribe     bool   `json:"can_subscribe"`
	LegacyDiscount   bool   `json:"legacy_discount"`
	ExternallyManaged bool  `json:"externally_managed"`
}

type CloudflareHost struct {
	Name    string `json:"Name"`
	Website string `json:"Website"`
}

type CloudflareMeta struct {
	PageRuleQuota      int  `json:"page_rule_quota"`
	WildcardProxiable  bool `json:"wildcard_proxiable"`
	PhishingDetected   bool `json:"phishing_detected"`
}

type CloudflareAccount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedOn string `json:"created_on"`
}

type CloudflareZone struct {
	ID                    string               `json:"id"`
	Name                  string               `json:"name"`
	DevelopmentMode       int                  `json:"development_mode"`
	OriginalNameServers   []string             `json:"original_name_servers,omitempty"`
	OriginalRegistrar     string               `json:"original_registrar,omitempty"`
	OriginalDNSHost       string               `json:"original_dnshost,omitempty"`
	CreatedOn             string               `json:"created_on,omitempty"`
	ModifiedOn            string               `json:"modified_on,omitempty"`
	NameServers           []string             `json:"name_servers,omitempty"`
	Owner                 CloudflareOwner      `json:"owner,omitempty"`
	Permissions           []string             `json:"permissions,omitempty"`
	Plan                  CloudflarePlan       `json:"plan,omitempty"`
	PlanPending           CloudflarePlan       `json:"plan_pending,omitempty"`
	Status                string               `json:"status"`
	Paused                bool                 `json:"paused"`
	Type                  string               `json:"type"`
	Host                  CloudflareHost       `json:"host,omitempty"`
	VanityNameServers     []string             `json:"vanity_name_servers,omitempty"`
	Betas                 []string             `json:"betas,omitempty"`
	DeactivationReason    string               `json:"deactivation_reason,omitempty"`
	Meta                  CloudflareMeta       `json:"meta,omitempty"`
	Account               CloudflareAccount    `json:"account,omitempty"`
	VerificationKey       string               `json:"verification_key,omitempty"`
}

type CloudflareDNSRecord struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Content    string  `json:"content"`
	TTL        int     `json:"ttl"`
	Proxied    *bool   `json:"proxied,omitempty"`
	ZoneID     string  `json:"zone_id"`
	ZoneName   string  `json:"zone_name"`
	CreatedOn  *string `json:"created_on,omitempty"`
	ModifiedOn *string `json:"modified_on,omitempty"`
}

type CloudflareLoadBalancer struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      *string  `json:"description,omitempty"`
	Enabled          bool     `json:"enabled"`
	TTL              int      `json:"ttl"`
	Proxied          bool     `json:"proxied"`
	DefaultPools     []string `json:"default_pools"`
	FallbackPool     *string  `json:"fallback_pool,omitempty"`
	SessionAffinity  string   `json:"session_affinity"`
	SteeringPolicy   string   `json:"steering_policy"`
}

type CloudflareOrigin struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Enabled bool   `json:"enabled"`
	Weight  *int   `json:"weight,omitempty"`
}

type CloudflarePool struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	Enabled            bool               `json:"enabled"`
	Healthy            bool               `json:"healthy"`
	MinimumOrigins     int                `json:"minimum_origins"`
	Origins            []CloudflareOrigin `json:"origins"`
	Monitor            *string            `json:"monitor,omitempty"`
	NotificationEmail  *string            `json:"notification_email,omitempty"`
}

type CloudflareWAFRule struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Expression  string  `json:"expression"`
	Action      string  `json:"action"`
	Enabled     bool    `json:"enabled"`
	Ref         *string `json:"ref,omitempty"`
	Version     *string `json:"version,omitempty"`
}

// GetCloudflareConfig retrieves the current Cloudflare configuration from environment variables
func (a *App) GetCloudflareConfig() CloudflareAPIConfig {
	accountName := os.Getenv("CLOUDFLARE_ACCOUNT_NAME")
	if accountName == "" {
		accountName = "Doctolib"
	}
	return CloudflareAPIConfig{
		Token:       os.Getenv("CLOUDFLARE_API_TOKEN"),
		AccountName: accountName,
	}
}

// SetCloudflareConfig stores the Cloudflare configuration in environment variables
func (a *App) SetCloudflareConfig(config CloudflareAPIConfig) error {
	if config.Token != "" {
		os.Setenv("CLOUDFLARE_API_TOKEN", config.Token)
	}
	if config.AccountName != "" {
		os.Setenv("CLOUDFLARE_ACCOUNT_NAME", config.AccountName)
	}
	return nil
}

// GetCloudflareZones retrieves all zones for the configured account
func (a *App) GetCloudflareZones(config CloudflareAPIConfig) ([]CloudflareZone, error) {
	if config.Token == "" || config.AccountName == "" {
		return nil, fmt.Errorf("cloudflare token and account name are required")
	}

	// Use the specific yak path that has cloudflare support
	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	
	// First check if cloudflare command is available
	checkCmd := exec.Command(yakPath, "cloudflare", "--help")
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("cloudflare command not available in yak CLI at %s. Error: %v", yakPath, err)
	}
	
	// Build command with authentication
	args := []string{"cloudflare", "zone", "list", "--json"}
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd := exec.Command(yakPath, args...)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %v. Command: %s", err, cmd.String())
	}

	// The actual JSON structure might be different, let's handle multiple possible formats
	var zones []CloudflareZone
	
	// First try direct array unmarshaling
	if err := json.Unmarshal(output, &zones); err != nil {
		// Try wrapped response format
		var response struct {
			Result []CloudflareZone `json:"result"`
			Success bool `json:"success"`
		}
		if err2 := json.Unmarshal(output, &response); err2 != nil {
			// Try data wrapper format
			var dataResponse struct {
				Data []CloudflareZone `json:"data"`
			}
			if err3 := json.Unmarshal(output, &dataResponse); err3 != nil {
				return nil, fmt.Errorf("failed to parse zones response. Tried multiple formats. Output: %s. Errors: direct=%v, result=%v, data=%v", string(output), err, err2, err3)
			}
			zones = dataResponse.Data
		} else {
			zones = response.Result
		}
	}

	return zones, nil
}

// GetCloudflareDNSRecords retrieves DNS records for a specific zone
func (a *App) GetCloudflareDNSRecords(config CloudflareAPIConfig, zone string) ([]CloudflareDNSRecord, error) {
	if config.Token == "" || config.AccountName == "" {
		return nil, fmt.Errorf("cloudflare token and account name are required")
	}
	if zone == "" {
		return nil, fmt.Errorf("zone is required")
	}

	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	
	// Build command with authentication
	args := []string{"cloudflare", "dns", "list", "--json", "--zone", zone}
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd := exec.Command(yakPath, args...)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %v. Command: %s", err, cmd.String())
	}

	// Handle multiple possible JSON response formats
	var records []CloudflareDNSRecord
	
	// First try direct array unmarshaling
	if err := json.Unmarshal(output, &records); err != nil {
		// Try wrapped response format
		var response struct {
			Result []CloudflareDNSRecord `json:"result"`
			Success bool `json:"success"`
		}
		if err2 := json.Unmarshal(output, &response); err2 != nil {
			// Try data wrapper format
			var dataResponse struct {
				Data []CloudflareDNSRecord `json:"data"`
			}
			if err3 := json.Unmarshal(output, &dataResponse); err3 != nil {
				return nil, fmt.Errorf("failed to parse DNS records response. Tried multiple formats. Output: %s. Errors: direct=%v, result=%v, data=%v", string(output), err, err2, err3)
			}
			records = dataResponse.Data
		} else {
			records = response.Result
		}
	}

	return records, nil
}

// CreateCloudflareDNSRecord creates a new DNS record
func (a *App) CreateCloudflareDNSRecord(config CloudflareAPIConfig, zone string, record map[string]interface{}) error {
	if config.Token == "" || config.AccountName == "" {
		return fmt.Errorf("cloudflare token and account name are required")
	}
	if zone == "" {
		return fmt.Errorf("zone is required")
	}

	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	args := []string{"cloudflare", "dns", "create", "--zone", zone}
	
	// Add record parameters
	if recordType, ok := record["type"].(string); ok {
		args = append(args, "--type", recordType)
	}
	if name, ok := record["name"].(string); ok {
		args = append(args, "--name", name)
	}
	if content, ok := record["content"].(string); ok {
		args = append(args, "--content", content)
	}
	if ttl, ok := record["ttl"].(float64); ok {
		args = append(args, "--ttl", fmt.Sprintf("%.0f", ttl))
	}
	if proxied, ok := record["proxied"].(bool); ok && proxied {
		args = append(args, "--proxied")
	}
	
	// Add authentication
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd := exec.Command(yakPath, args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create DNS record: %v", err)
	}

	return nil
}

// DeleteCloudflareDNSRecord deletes a DNS record
func (a *App) DeleteCloudflareDNSRecord(config CloudflareAPIConfig, zone string, recordID string) error {
	if config.Token == "" || config.AccountName == "" {
		return fmt.Errorf("cloudflare token and account name are required")
	}
	if zone == "" || recordID == "" {
		return fmt.Errorf("zone and record ID are required")
	}

	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	args := []string{"cloudflare", "dns", "delete", "--zone", zone, "--record-id", recordID}
	
	// Add authentication
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd := exec.Command(yakPath, args...)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete DNS record: %v", err)
	}

	return nil
}

// GetCloudflareLoadBalancers retrieves all load balancers for a specific zone
func (a *App) GetCloudflareLoadBalancers(config CloudflareAPIConfig, zone string) ([]CloudflareLoadBalancer, error) {
	if config.Token == "" || config.AccountName == "" {
		return nil, fmt.Errorf("cloudflare token and account name are required")
	}
	if zone == "" {
		return nil, fmt.Errorf("zone is required for load balancers")
	}

	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	
	// Build command with authentication
	args := []string{"cloudflare", "lb", "list", "--json", "--zone", zone}
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd := exec.Command(yakPath, args...)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get load balancers: %v. Command: %s", err, cmd.String())
	}

	// Handle multiple possible JSON response formats
	var lbs []CloudflareLoadBalancer
	
	// First try direct array unmarshaling
	if err := json.Unmarshal(output, &lbs); err != nil {
		// Try wrapped response format
		var response struct {
			Result []CloudflareLoadBalancer `json:"result"`
			Success bool `json:"success"`
		}
		if err2 := json.Unmarshal(output, &response); err2 != nil {
			// Try data wrapper format
			var dataResponse struct {
				Data []CloudflareLoadBalancer `json:"data"`
			}
			if err3 := json.Unmarshal(output, &dataResponse); err3 != nil {
				return nil, fmt.Errorf("failed to parse load balancers response. Tried multiple formats. Output: %s. Errors: direct=%v, result=%v, data=%v", string(output), err, err2, err3)
			}
			lbs = dataResponse.Data
		} else {
			lbs = response.Result
		}
	}

	return lbs, nil
}

// GetCloudflarePools retrieves all pools for the account  
func (a *App) GetCloudflarePools(config CloudflareAPIConfig) ([]CloudflarePool, error) {
	if config.Token == "" || config.AccountName == "" {
		return nil, fmt.Errorf("cloudflare token and account name are required")
	}

	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	
	// Build command with authentication
	args := []string{"cloudflare", "pool", "list", "--json"}
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd := exec.Command(yakPath, args...)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pools: %v", err)
	}

	var pools []CloudflarePool
	if err := json.Unmarshal(output, &pools); err != nil {
		return nil, fmt.Errorf("failed to parse pools response: %v", err)
	}

	return pools, nil
}

// GetCloudflareWAFRules retrieves WAF rules for a specific zone
func (a *App) GetCloudflareWAFRules(config CloudflareAPIConfig, zone string) ([]CloudflareWAFRule, error) {
	if config.Token == "" || config.AccountName == "" {
		return nil, fmt.Errorf("cloudflare token and account name are required")
	}
	if zone == "" {
		return nil, fmt.Errorf("zone is required")
	}

	yakPath := "/Users/sergiosantiago/projects/doctolib/yak_fix/yak"
	
	// Try rulesets first, then fall back to rules if that doesn't work
	var cmd *exec.Cmd
	var cmdType string
	
	// WAF rules require 'list' subcommand
	args := []string{"cloudflare", "waf", "rules", "list", "--json", "--zone", zone}
	if config.Token != "" {
		args = append(args, "--token", config.Token)
	}
	if config.AccountName != "" {
		args = append(args, "--account-name", config.AccountName)
	}
	
	cmd = exec.Command(yakPath, args...)
	cmdType = "rules"
	
	output, err := cmd.Output()
	if err != nil {
		// Try rulesets list if rules fail
		args = []string{"cloudflare", "waf", "rulesets", "list", "--json", "--zone", zone}
		if config.Token != "" {
			args = append(args, "--token", config.Token)
		}
		if config.AccountName != "" {
			args = append(args, "--account-name", config.AccountName)
		}
		
		// Fall back to rulesets
		cmd = exec.Command(yakPath, args...)
		cmdType = "rulesets"
		
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get WAF %s: %v", cmdType, err)
		}
	}

	var rules []CloudflareWAFRule
	if err := json.Unmarshal(output, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse WAF %s response: %v", cmdType, err)
	}

	return rules, nil
}

// ResolveHostname performs simple hostname to IP resolution
func (a *App) ResolveHostname(hostname string) (string, error) {
	if hostname == "" {
		return "", fmt.Errorf("hostname is required")
	}

	// Trim any whitespace
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return "", fmt.Errorf("hostname cannot be empty")
	}

	// Simple IP lookup with short timeout
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", fmt.Errorf("could not resolve %s: %v", hostname, err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for %s", hostname)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Resolved IPs for %s:\n", hostname))
	
	for _, ip := range ips {
		result.WriteString(fmt.Sprintf("  %s\n", ip.String()))
	}

	return result.String(), nil
}