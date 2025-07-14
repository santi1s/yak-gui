package main

import (
	"fmt"
	"os/exec"
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

// Helper function to get nested string values from JSON
func getNestedString(data map[string]interface{}, keys ...string) string {
	current := data
	for _, key := range keys {
		if val, ok := current[key]; ok {
			if len(keys) == 1 {
				// This is the last key, return the string value
				if str, ok := val.(string); ok {
					return str
				}
			} else {
				// More keys to traverse
				if nested, ok := val.(map[string]interface{}); ok {
					current = nested
					keys = keys[1:]
					return getNestedString(current, keys...)
				}
			}
		}
		return ""
	}
	return ""
}

// Helper function to extract conditions from ArgoCD app status
func extractConditions(data map[string]interface{}) []string {
	var conditions []string
	
	if status, ok := data["status"].(map[string]interface{}); ok {
		if conditionsArray, ok := status["conditions"].([]interface{}); ok {
			for _, conditionInterface := range conditionsArray {
				if condition, ok := conditionInterface.(map[string]interface{}); ok {
					if condType := getString(condition, "type"); condType != "" {
						if message := getString(condition, "message"); message != "" {
							conditions = append(conditions, fmt.Sprintf("%s: %s", condType, message))
						} else {
							conditions = append(conditions, condType)
						}
					}
				}
			}
		}
	}
	
	return conditions
}

// Helper function to check if an ArgoCD app is suspended
func isAppSuspended(data map[string]interface{}) bool {
	// Check if sync policy is disabled or if there's no automated sync
	if spec, ok := data["spec"].(map[string]interface{}); ok {
		if syncPolicy, ok := spec["syncPolicy"].(map[string]interface{}); ok {
			// If automated sync is not present, it might be manually managed
			if automated, ok := syncPolicy["automated"]; !ok || automated == nil {
				return true
			}
		}
	}
	return false
}

// Helper function to extract sync loop information
func extractSyncLoop(data map[string]interface{}) string {
	// Check operation state for sync loop info
	if status, ok := data["status"].(map[string]interface{}); ok {
		if operationState, ok := status["operationState"].(map[string]interface{}); ok {
			if phase := getString(operationState, "phase"); phase != "" {
				return phase
			}
		}
		
		// Fallback to sync status
		if sync, ok := status["sync"].(map[string]interface{}); ok {
			if status := getString(sync, "status"); status != "" {
				return status
			}
		}
	}
	return ""
}

