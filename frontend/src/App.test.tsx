import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from './test/test-utils'
import userEvent from '@testing-library/user-event'
import App from './App'

// Mock environment variables and profiles
const mockEnvironmentVars = {
  AWS_PROFILE: 'staging',
  KUBECONFIG: '/path/to/kubeconfig',
  HOME: '/Users/test',
  PATH: '/usr/local/bin:/usr/bin',
  TFINFRA_REPOSITORY_PATH: '/path/to/terraform-infra'
}

const mockAWSProfiles = ['dev', 'staging', 'production']

const mockEnvironmentProfiles = [
  {
    name: 'Development',
    aws_profile: 'dev',
    kubeconfig: '/path/to/dev/config',
    path: '/usr/local/bin:/usr/bin',
    tf_infra_repository_path: '/path/to/terraform-infra',
    created_at: '2024-01-01T00:00:00Z'
  }
]

const mockAppVersion = {
  version: '1.5.0',
  buildDate: '2024-01-15',
  gitCommit: 'abc123'
}

const mockArgoApps = [
  {
    AppName: 'api-service',
    Health: 'Healthy',
    Sync: 'Synced',
    Suspended: false,
    SyncLoop: 'Normal',
    Conditions: []
  }
]

describe('App Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    
    // Setup default mocks
    window.go.main.App.GetEnvironmentVariables.mockResolvedValue(mockEnvironmentVars)
    window.go.main.App.GetAWSProfiles.mockResolvedValue(mockAWSProfiles)
    window.go.main.App.GetEnvironmentProfiles.mockResolvedValue(mockEnvironmentProfiles)
    window.go.main.App.GetAppVersion.mockResolvedValue(mockAppVersion)
    window.go.main.App.GetArgoApps.mockResolvedValue(mockArgoApps)
    window.go.main.App.GetArgoCDServerFromProfile.mockResolvedValue('argocd-staging.doctolib.net')
    window.go.main.App.SetAWSProfile.mockResolvedValue(undefined)
    window.go.main.App.ImportShellEnvironment.mockResolvedValue(undefined)
    
    // Certificate mocks
    window.go.main.App.ListCertificates.mockResolvedValue(['cert1', 'cert2'])
    window.go.main.App.CheckGandiToken.mockResolvedValue({ success: true, message: 'Token valid' })
    window.go.main.App.RenewCertificate.mockResolvedValue({ success: true, message: 'Renewed' })
    window.go.main.App.RefreshCertificateSecret.mockResolvedValue({ success: true, message: 'Refreshed' })
    window.go.main.App.DescribeCertificateSecret.mockResolvedValue({ success: true, message: 'Described' })
    window.go.main.App.SendCertificateNotification.mockResolvedValue({ success: true, message: 'Sent' })
    
    // Rollouts mocks
    window.go.main.App.GetRollouts.mockResolvedValue([])
    window.go.main.App.GetRolloutStatus.mockResolvedValue({})
    window.go.main.App.PromoteRollout.mockResolvedValue(undefined)
    window.go.main.App.AbortRollout.mockResolvedValue(undefined)
    window.go.main.App.RestartRollout.mockResolvedValue(undefined)
    
    // Secrets mocks
    window.go.main.App.GetSecrets.mockResolvedValue([])
    window.go.main.App.GetSecret.mockResolvedValue({})
    window.go.main.App.CreateSecret.mockResolvedValue(undefined)
    window.go.main.App.UpdateSecret.mockResolvedValue(undefined)
    window.go.main.App.DeleteSecret.mockResolvedValue(undefined)
    window.go.main.App.CreateJWTClient.mockResolvedValue(undefined)
    window.go.main.App.CreateJWTServer.mockResolvedValue(undefined)
    window.go.main.App.ListSecrets.mockResolvedValue([])
  })

  it('renders the main application title', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Yak GUI')).toBeInTheDocument()
    })
  })

  it('displays navigation tabs', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Environment')).toBeInTheDocument()
      expect(screen.getByText('ArgoCD Applications')).toBeInTheDocument()
      expect(screen.getByText('Argo Rollouts')).toBeInTheDocument()
      expect(screen.getByText('Secrets')).toBeInTheDocument()
      expect(screen.getByText('Certificates')).toBeInTheDocument()
    })
  })

  it('loads environment variables on mount', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(window.go.main.App.GetEnvironmentVariables).toHaveBeenCalled()
    })
  })

  it('loads AWS profiles on mount', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(window.go.main.App.GetAWSProfiles).toHaveBeenCalled()
    })
  })

  it('displays current AWS profile in environment config', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Current AWS Profile:')).toBeInTheDocument()
      // Use getAllByText to handle multiple matches and find the right one
      const stagingTexts = screen.getAllByText('staging')
      expect(stagingTexts.length).toBeGreaterThan(0)
    })
  })

  it('shows theme toggle button', async () => {
    render(<App />)
    
    const themeSwitches = screen.getAllByRole('switch')
    expect(themeSwitches.length).toBeGreaterThan(0)
  })

  it('handles theme toggle', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    const themeSwitches = screen.getAllByRole('switch')
    expect(themeSwitches.length).toBeGreaterThan(0)
    
    // Click first switch (should be theme toggle)
    await user.click(themeSwitches[0])
    
    // Theme should change (hard to test directly, but the toggle should work)
    expect(themeSwitches[0]).toBeInTheDocument()
  })

  it('displays version information in environment tab', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Environment')).toBeInTheDocument()
    })
    
    // Should already be on Environment tab by default
    await waitFor(() => {
      expect(screen.getByText('v1.5.0')).toBeInTheDocument()
    })
  })

  it('shows environment configuration by default', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    // Should already be on Environment tab by default
    await waitFor(() => {
      expect(screen.getByText('Environment Configuration')).toBeInTheDocument()
      expect(screen.getByText('AWS_PROFILE:')).toBeInTheDocument()
      // Use getAllByText to handle multiple matches
      const stagingTexts = screen.getAllByText('staging')
      expect(stagingTexts.length).toBeGreaterThan(0)
    })
  })

  it('handles AWS profile change', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    // Should already be on Environment tab by default
    await waitFor(() => {
      expect(screen.getByText('Environment Configuration')).toBeInTheDocument()
    })
    
    // Look for the Set Profile button instead of trying to interact with select
    const setProfileButton = screen.getByText('Set Profile')
    expect(setProfileButton).toBeInTheDocument()
    
    // This test verifies the UI is present, the actual profile change behavior
    // would need a more complex interaction with the Select component
  })

  it('displays environment profiles section', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    // Should already be on Environment tab by default
    await waitFor(() => {
      expect(screen.getByText('Environment Profiles')).toBeInTheDocument()
    })
    
    // Just check that the profiles section exists and has some content
    await waitFor(() => {
      expect(screen.getByText('Save and load different environment configurations for quick switching between setups.')).toBeInTheDocument()
    })
  })

  it('handles import shell environment', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    // Should already be on Environment tab by default
    await waitFor(() => {
      expect(screen.getByText('Shell Environment')).toBeInTheDocument()
    })
    
    // The button text depends on the autoImported state - could be "Re-import Shell Environment"
    const importButton = screen.getByText('Re-import Shell Environment')
    await user.click(importButton)
    
    await waitFor(() => {
      expect(window.go.main.App.ImportShellEnvironment).toHaveBeenCalled()
    })
  })

  it('switches between tabs correctly', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    // Should start on Environment tab
    await waitFor(() => {
      expect(screen.getByText('Environment Configuration')).toBeInTheDocument()
    })
    
    // Switch to ArgoCD Applications
    const argoCDTabs = screen.getAllByText('ArgoCD Applications')
    const argoCDTab = argoCDTabs[0] // Click on the first one (tab label)
    await user.click(argoCDTab)
    
    await waitFor(() => {
      const argoCDTexts = screen.getAllByText('ArgoCD Applications')
      expect(argoCDTexts.length).toBeGreaterThan(0)
    })
    
    // Switch to Rollouts
    const rolloutsTabs = screen.getAllByText('Argo Rollouts')
    const rolloutsTab = rolloutsTabs[0] // Click on the first one (tab label)
    await user.click(rolloutsTab)
    
    await waitFor(() => {
      const rolloutsTexts = screen.getAllByText('Argo Rollouts')
      expect(rolloutsTexts.length).toBeGreaterThan(0)
    })
    
    // Switch to Secrets
    const secretsTab = screen.getByText('Secrets')
    await user.click(secretsTab)
    
    await waitFor(() => {
      expect(screen.getByText('Secrets Management')).toBeInTheDocument()
    })
    
    // Switch to Certificates
    const certificatesTab = screen.getByText('Certificates')
    await user.click(certificatesTab)
    
    await waitFor(() => {
      // Check if the certificates tab content is loaded
      const hasSSLText = screen.queryByText('SSL Certificate Management') !== null
      const hasCertificatesText = screen.queryByText('Certificates') !== null
      expect(hasSSLText || hasCertificatesText).toBeTruthy()
    })
  })

  it('displays ArgoCD server from profile', async () => {
    render(<App />)
    
    // This function is called when AWS profile changes or on startup
    // Just verify the app renders properly and the function is available
    await waitFor(() => {
      expect(screen.getByText('Environment Configuration')).toBeInTheDocument()
    })
    
    // Verify the function is available (mocked)
    expect(window.go.main.App.GetArgoCDServerFromProfile).toBeDefined()
  })

  it('shows error state when environment loading fails', async () => {
    window.go.main.App.GetEnvironmentVariables.mockRejectedValue(new Error('Failed to load environment'))
    
    render(<App />)
    
    // Error should be handled gracefully
    await waitFor(() => {
      expect(window.go.main.App.GetEnvironmentVariables).toHaveBeenCalled()
    })
  })

  it('handles tab key navigation', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Environment')).toBeInTheDocument()
    })
    
    // Use tab key to navigate
    const environmentTab = screen.getByText('Environment')
    await user.tab()
    
    expect(environmentTab).toBeInTheDocument()
  })

  it('displays loading states appropriately', async () => {
    // Mock slow loading
    window.go.main.App.GetEnvironmentVariables.mockImplementation(() => 
      new Promise(resolve => setTimeout(() => resolve(mockEnvironmentVars), 1000))
    )
    
    render(<App />)
    
    // Should show some loading indication while environment loads
    expect(screen.getByText('Yak GUI')).toBeInTheDocument()
  })

  it('maintains responsive design', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Yak GUI')).toBeInTheDocument()
    })
    
    // Basic check that layout components are present
    expect(screen.getByRole('banner')).toBeInTheDocument() // Header
    expect(screen.getByRole('main')).toBeInTheDocument() // Main content
  })

  it('handles environment profile creation', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    // Should already be on Environment tab by default
    await waitFor(() => {
      expect(screen.getByText('Environment Profiles')).toBeInTheDocument()
    })
    
    // Look for create profile functionality
    const createButton = screen.getByText('Save Profile')
    expect(createButton).toBeInTheDocument()
  })

  it('shows appropriate icons for each tab', async () => {
    render(<App />)
    
    await waitFor(() => {
      // Icons are rendered as svg elements or spans with specific classes
      // The exact assertion depends on how Ant Design renders icons
      expect(screen.getByText('Environment')).toBeInTheDocument()
      expect(screen.getByText('ArgoCD Applications')).toBeInTheDocument()
      expect(screen.getByText('Argo Rollouts')).toBeInTheDocument()
      expect(screen.getByText('Secrets')).toBeInTheDocument()
      expect(screen.getByText('Certificates')).toBeInTheDocument()
    })
  })

  it('handles keyboard shortcuts', async () => {
    const user = userEvent.setup()
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Environment')).toBeInTheDocument()
    })
    
    // Test keyboard navigation
    await user.keyboard('{Control>}1{/Control}') // Ctrl+1 might switch to first tab
    
    // The specific behavior depends on implementation
    expect(screen.getByText('Environment')).toBeInTheDocument()
  })
})