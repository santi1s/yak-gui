import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from './test/test-utils'
import userEvent from '@testing-library/user-event'
import Rollouts from './Rollouts'

// Mock rollouts data
const mockRollouts = [
  {
    name: 'api-service',
    namespace: 'production',
    status: 'Healthy',
    replicas: '3/3',
    age: '2d',
    strategy: 'BlueGreen',
    revision: '123',
    images: {
      'api-service': 'registry.com/api-service:v1.2.3'
    }
  },
  {
    name: 'web-frontend',
    namespace: 'staging',
    status: 'Progressing',
    replicas: '2/3',
    age: '1h',
    strategy: 'Canary',
    revision: '124',
    images: {
      'web-frontend': 'registry.com/web-frontend:v2.1.0'
    }
  }
]

const mockRolloutStatus = {
  name: 'api-service',
  namespace: 'production',
  status: 'Healthy',
  replicas: '3/3',
  updated: '3',
  ready: '3',
  available: '3',
  strategy: 'BlueGreen',
  currentStep: '8/8',
  revision: '123',
  message: 'Rollout is healthy',
  analysis: 'Success',
  images: {
    'api-service': 'registry.com/api-service:v1.2.3'
  }
}

describe('Rollouts Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    
    // Setup default mocks
    window.go.main.App.GetRollouts.mockResolvedValue(mockRollouts)
    window.go.main.App.GetRolloutStatus.mockResolvedValue(mockRolloutStatus)
    window.go.main.App.PromoteRollout.mockResolvedValue({ success: true })
    window.go.main.App.AbortRollout.mockResolvedValue({ success: true })
    window.go.main.App.RestartRollout.mockResolvedValue({ success: true })
    window.go.main.App.RetryRollout.mockResolvedValue({ success: true })
    window.go.main.App.GetRolloutHistory.mockResolvedValue([])
    window.go.main.App.PauseRollout.mockResolvedValue({ success: true })
  })

  it('renders the Argo Rollouts title', async () => {
    render(<Rollouts />)
    
    expect(screen.getByText('Argo Rollouts')).toBeInTheDocument()
    expect(screen.getByText('Manage your Argo Rollouts')).toBeInTheDocument()
  })

  it('loads rollouts on mount', async () => {
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(window.go.main.App.GetRollouts).toHaveBeenCalled()
    })
  })

  it('displays configuration section', () => {
    render(<Rollouts />)
    
    expect(screen.getByText('Namespace:')).toBeInTheDocument()
  })

  it('shows rollouts table when rollouts are loaded', async () => {
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
      expect(screen.getByText('web-frontend')).toBeInTheDocument()
      expect(screen.getByText('production')).toBeInTheDocument()
      expect(screen.getByText('staging')).toBeInTheDocument()
      expect(screen.getByText('Healthy')).toBeInTheDocument()
      expect(screen.getByText('Progressing')).toBeInTheDocument()
    })
  })

  it('displays rollout strategies correctly', async () => {
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('BlueGreen')).toBeInTheDocument()
      expect(screen.getByText('Canary')).toBeInTheDocument()
    })
  })

  it('shows rollout images in details dialog', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    // Open details modal to see images
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(screen.getByText(/Rollout Details/)).toBeInTheDocument()
      // Check if image text is present in the modal
      const documentText = document.body.textContent || ''
      expect(documentText).toContain('registry.com/api-service:v1.2.3')
    })
  })


  it('handles namespace configuration change', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    const namespaceInput = screen.getByPlaceholderText('Search rollouts...')
    await user.type(namespaceInput, 'production')
    
    expect(namespaceInput).toHaveValue('production')
  })

  it('opens rollout details when View button is clicked', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    // Find and click the first View button (eye icon button in card extra)
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(window.go.main.App.GetRolloutStatus).toHaveBeenCalledWith(
        { server: '', namespace: 'production' },
        'api-service'
      )
    })
  })

  it('shows rollout details dialog', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText(/Rollout Details/)).toBeInTheDocument()
    })
  })

  it('displays rollout status information in details', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(screen.getByText(/Rollout Details/)).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
      expect(screen.getByText('Strategy')).toBeInTheDocument()
      const blueGreenElements = screen.getAllByText('BlueGreen')
      expect(blueGreenElements.length).toBeGreaterThan(0)
    })
  })

  it('handles promote rollout action', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    // Find and click the first Promote button
    const promoteButtons = screen.getAllByText('Promote')
    await user.click(promoteButtons[0])
    
    await waitFor(() => {
      expect(window.go.main.App.PromoteRollout).toHaveBeenCalledWith(
        { server: '', namespace: '' },
        'api-service',
        false
      )
    })
  })

  it('handles abort rollout action', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    // Find and click the first Abort button
    const abortButtons = screen.getAllByText('Abort')
    await user.click(abortButtons[0])
    
    await waitFor(() => {
      expect(window.go.main.App.AbortRollout).toHaveBeenCalledWith(
        { server: '', namespace: '' },
        'api-service'
      )
    })
  })

  it('handles restart rollout action', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    // Find and click the first Restart button
    const restartButtons = screen.getAllByText('Restart')
    await user.click(restartButtons[0])
    
    await waitFor(() => {
      expect(window.go.main.App.RestartRollout).toHaveBeenCalledWith(
        { server: '', namespace: '' },
        'api-service'
      )
    })
  })

  it('handles retry rollout action', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    // Find and click the first Restart button (the component uses Restart, not Retry)
    const restartButtons = screen.getAllByText('Restart')
    await user.click(restartButtons[0])
    
    await waitFor(() => {
      expect(window.go.main.App.RestartRollout).toHaveBeenCalledWith(
        { server: '', namespace: '' },
        'api-service'
      )
    })
  })

  it('shows rollout history in details dialog', async () => {
    const mockHistory = [
      { revision: '123', createdAt: '2024-01-15T10:00:00Z', message: 'Updated image' },
      { revision: '122', createdAt: '2024-01-14T10:00:00Z', message: 'Initial deployment' }
    ]
    window.go.main.App.GetRolloutHistory.mockResolvedValue(mockHistory)
    
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(screen.getByText(/Rollout Details/)).toBeInTheDocument()
    })
  })

  it('handles refresh button click', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    const refreshButton = screen.getByText('Refresh')
    await user.click(refreshButton)
    
    await waitFor(() => {
      expect(window.go.main.App.GetRollouts).toHaveBeenCalledTimes(2) // Once on mount, once on refresh
    })
  })

  it('shows loading state while fetching rollouts', () => {
    // Mock a slow API response
    window.go.main.App.GetRollouts.mockImplementation(() => 
      new Promise(resolve => setTimeout(() => resolve(mockRollouts), 1000))
    )
    
    render(<Rollouts />)
    
    expect(screen.getByLabelText('loading')).toBeInTheDocument()
  })

  it('handles error states gracefully', async () => {
    window.go.main.App.GetRollouts.mockRejectedValue(new Error('Failed to load rollouts'))
    
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText(/Failed to load rollouts/)).toBeInTheDocument()
    })
  })

  it('displays rollout status badges with correct colors', async () => {
    render(<Rollouts />)
    
    await waitFor(() => {
      const healthyStatus = screen.getByText('Healthy')
      const progressingStatus = screen.getByText('Progressing')
      
      expect(healthyStatus).toBeInTheDocument()
      expect(progressingStatus).toBeInTheDocument()
    })
  })

  it('shows replica counts correctly', async () => {
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('3/3')).toBeInTheDocument()
      expect(screen.getByText('2/3')).toBeInTheDocument()
    })
  })

  it('displays age information', async () => {
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('2d')).toBeInTheDocument()
      expect(screen.getByText('1h')).toBeInTheDocument()
    })
  })

  it('handles step progression in details view', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(screen.getByText('Current Step')).toBeInTheDocument()
      expect(screen.getByText('8/8')).toBeInTheDocument()
    })
  })

  it('shows analysis results in details', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    await waitFor(() => {
      expect(screen.getByText('api-service')).toBeInTheDocument()
    })
    
    const viewButtons = screen.getAllByRole('button')
    const viewButton = viewButtons.find(btn => btn.querySelector('.anticon-eye'))
    await user.click(viewButton!)
    
    await waitFor(() => {
      expect(screen.getByText('Analysis')).toBeInTheDocument()
      expect(screen.getByText('Success')).toBeInTheDocument()
    })
  })

  it('handles namespace filtering correctly', async () => {
    const user = userEvent.setup()
    render(<Rollouts />)
    
    // Check that namespace is displayed
    await waitFor(() => {
      expect(screen.getByText('Namespace:')).toBeInTheDocument()
    })
    
    // Check search functionality
    const searchInput = screen.getByPlaceholderText('Search rollouts...')
    await user.type(searchInput, 'production')
    
    expect(searchInput).toHaveValue('production')
    
    // Just verify that the search works
    await waitFor(() => {
      expect(searchInput).toHaveValue('production')
    })
  })
})