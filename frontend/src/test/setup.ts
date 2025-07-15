import '@testing-library/jest-dom'

// Mock window.matchMedia for Ant Design responsive components
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock getComputedStyle for Ant Design modals
Object.defineProperty(window, 'getComputedStyle', {
  value: vi.fn().mockImplementation(() => ({
    getPropertyValue: vi.fn(() => ''),
    width: '1024px',
    height: '768px',
    overflow: 'visible',
    position: 'static',
    zIndex: 'auto',
  })),
})

// Mock scrollTo
Object.defineProperty(window, 'scrollTo', {
  value: vi.fn(),
})

// Mock window.go for Wails bindings
global.window.go = {
  main: {
    App: {
      // ArgoCD mock functions
      GetArgoApps: vi.fn(),
      GetArgoAppDetail: vi.fn(),
      SyncArgoApp: vi.fn(),
      RefreshArgoApp: vi.fn(),
      SuspendArgoApp: vi.fn(),
      UnsuspendArgoApp: vi.fn(),
      LoginToArgoCD: vi.fn(),
      GetArgoCDServerFromProfile: vi.fn(),
      
      // Rollouts mock functions
      GetRollouts: vi.fn(),
      GetRolloutStatus: vi.fn(),
      GetRolloutHistory: vi.fn(),
      PromoteRollout: vi.fn(),
      AbortRollout: vi.fn(),
      RestartRollout: vi.fn(),
      RetryRollout: vi.fn(),
      PauseRollout: vi.fn(),
      
      // Secrets mock functions
      GetSecrets: vi.fn(),
      GetSecret: vi.fn(),
      CreateSecret: vi.fn(),
      UpdateSecret: vi.fn(),
      DeleteSecret: vi.fn(),
      CreateJWTClient: vi.fn(),
      CreateJWTServer: vi.fn(),
      ListSecrets: vi.fn(),
      
      // Certificates mock functions
      CheckGandiToken: vi.fn(),
      RenewCertificate: vi.fn(),
      RefreshCertificateSecret: vi.fn(),
      DescribeCertificateSecret: vi.fn(),
      SendCertificateNotification: vi.fn(),
      ListCertificates: vi.fn(),
      
      // Environment mock functions
      GetCurrentAWSProfile: vi.fn(),
      SetAWSProfile: vi.fn(),
      GetAWSProfiles: vi.fn(),
      GetEnvironmentVariables: vi.fn(),
      ImportShellEnvironment: vi.fn(),
      SaveEnvironmentProfile: vi.fn(),
      LoadEnvironmentProfile: vi.fn(),
      DeleteEnvironmentProfile: vi.fn(),
      GetEnvironmentProfiles: vi.fn(),
      GetShellPATH: vi.fn(),
      GetShellEnvironment: vi.fn(),
      SetKubeconfig: vi.fn(),
      SetPATH: vi.fn(),
      SetTfInfraRepositoryPath: vi.fn(),
      
      // Utility mock functions
      GetAppVersion: vi.fn(),
      TestSimpleArray: vi.fn(),
    }
  }
}

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockImplementation(() => Promise.resolve()),
  },
})

// Mock notification API
global.Notification = {
  permission: 'granted',
  requestPermission: vi.fn(() => Promise.resolve('granted')),
} as any

// Suppress React act() warnings in tests
const originalError = console.error
beforeEach(() => {
  console.error = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Warning: An update to') &&
      args[0].includes('was not wrapped in act')
    ) {
      return
    }
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Warning: [antd: Menu]') &&
      args[0].includes('is deprecated')
    ) {
      return
    }
    originalError.call(console, ...args)
  }
})

afterEach(() => {
  console.error = originalError
})