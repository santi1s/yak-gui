// Feature flags configuration for controlling tab visibility and experimental features
export interface FeatureFlags {
  // Tab visibility flags
  showEnvironmentTab: boolean;
  showArgoCDTab: boolean;
  showRolloutsTab: boolean;
  showSecretsTab: boolean;
  showCertificatesTab: boolean;
  showTFETab: boolean;
  
  // Experimental features
  enableAutoRefresh: boolean;
  enableDarkMode: boolean;
  enableDetailedLogging: boolean;
}

// Default feature flags - these are the flags that are enabled by default
export const defaultFeatureFlags: FeatureFlags = {
  // Core tabs - always enabled
  showEnvironmentTab: true,
  showArgoCDTab: true,
  showRolloutsTab: true,
  showSecretsTab: true,
  showCertificatesTab: true,
  
  // New experimental tabs - disabled and not configurable
  showTFETab: false,
  
  // Feature flags for functionality
  enableAutoRefresh: true,
  enableDarkMode: true,
  enableDetailedLogging: false,
};

// Environment variable overrides for feature flags
// This allows you to override feature flags via environment variables
export const getFeatureFlagsFromEnv = (): Partial<FeatureFlags> => {
  const envFlags: Partial<FeatureFlags> = {};
  
  // Check for environment variables that override feature flags
  if (typeof window !== 'undefined' && window.go?.main?.App?.GetEnvironmentVariables) {
    // Note: This would need to be called asynchronously in practice
    // For now, we'll use a synchronous approach with localStorage for development
  }
  
  // Local storage overrides for development
  if (typeof localStorage !== 'undefined') {
    const localFlags = localStorage.getItem('yakgui_feature_flags');
    if (localFlags) {
      try {
        const parsed = JSON.parse(localFlags);
        return { ...envFlags, ...parsed };
      } catch (e) {
        console.warn('Failed to parse feature flags from localStorage:', e);
      }
    }
  }
  
  return envFlags;
};

// Main function to get current feature flags
export const getFeatureFlags = (): FeatureFlags => {
  const envOverrides = getFeatureFlagsFromEnv();
  return {
    ...defaultFeatureFlags,
    ...envOverrides,
    // Force certain tabs to their intended state regardless of overrides
    showEnvironmentTab: true,  // Always enabled
    showTFETab: false,         // Always disabled
  };
};

// Function to update feature flags (for development/testing)
export const updateFeatureFlags = (updates: Partial<FeatureFlags>): void => {
  if (typeof localStorage !== 'undefined') {
    const current = getFeatureFlagsFromEnv();
    const updated = { ...current, ...updates };
    localStorage.setItem('yakgui_feature_flags', JSON.stringify(updated));
    
    // Trigger a storage event to notify other parts of the app
    window.dispatchEvent(new StorageEvent('storage', {
      key: 'yakgui_feature_flags',
      newValue: JSON.stringify(updated),
    }));
  }
};

// Hook for React components to use feature flags
export const useFeatureFlags = (): [FeatureFlags, (updates: Partial<FeatureFlags>) => void] => {
  const [flags, setFlags] = React.useState<FeatureFlags>(getFeatureFlags);
  
  React.useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'yakgui_feature_flags') {
        setFlags(getFeatureFlags());
      }
    };
    
    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);
  
  const updateFlags = React.useCallback((updates: Partial<FeatureFlags>) => {
    updateFeatureFlags(updates);
    setFlags(getFeatureFlags());
  }, []);
  
  return [flags, updateFlags];
};

// React import for the hook
import React from 'react';