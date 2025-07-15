# Frontend Testing Setup

This project uses **Vitest** with **React Testing Library** for comprehensive frontend testing.

## 🧪 Testing Framework

- **Vitest**: Fast unit test framework built on top of Vite
- **React Testing Library**: Testing utilities for React components
- **Jest DOM**: Custom Jest matchers for DOM assertions
- **User Event**: Utilities for simulating user interactions

## 📁 Test Structure

```
frontend/src/
├── test/
│   ├── setup.ts          # Global test setup and mocks
│   ├── test-utils.tsx    # Custom render function with providers
│   └── README.md         # This file
├── App-antd.test.tsx     # Main application tests
├── Certificates.test.tsx # SSL certificate management tests
├── Secrets-antd.test.tsx # Secret management tests
└── Rollouts-antd.test.tsx # Argo Rollouts tests
```

## 🚀 Running Tests

```bash
# Run tests once
npm test

# Run tests in watch mode
npm run test:ui

# Run tests with coverage
npm run test:coverage
```

## 🔧 Configuration

### Vitest Config (`vitest.config.ts`)
- Uses jsdom environment for DOM testing
- Includes CSS processing
- Configures coverage reporting
- Sets up global test utilities

### Global Setup (`test/setup.ts`)
- Mocks Wails Go bindings (`window.go`)
- Mocks browser APIs (clipboard, notifications)
- Imports Jest DOM matchers

### Test Utils (`test/test-utils.tsx`)
- Provides custom render function with Ant Design ConfigProvider
- Exports all React Testing Library utilities

## 🎯 Test Coverage

Tests cover:

### **App Component**
- Navigation between tabs
- Theme switching
- Environment configuration
- AWS profile management
- Environment profiles
- Version information display

### **Certificates Component**
- SSL certificate management workflow
- Modal forms for operations
- GANDI token validation
- Copy-to-clipboard functionality
- Domain validation warnings
- Process step tracking

### **Secrets Component**
- Secret listing and navigation
- CRUD operations
- JWT client/server creation
- Breadcrumb navigation
- Configuration management
- Error handling

### **Rollouts Component**
- Argo Rollouts management
- Rollout operations (promote, abort, restart, retry)
- Status monitoring
- History tracking
- Configuration management

## 🛠️ Writing Tests

### Basic Test Structure
```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from './test/test-utils'
import userEvent from '@testing-library/user-event'
import ComponentToTest from './ComponentToTest'

describe('ComponentToTest', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Setup mocks
  })

  it('should render correctly', async () => {
    render(<ComponentToTest />)
    
    expect(screen.getByText('Expected Text')).toBeInTheDocument()
  })
})
```

### Mocking Wails Functions
```typescript
beforeEach(() => {
  window.go.main.App.SomeFunction.mockResolvedValue({ success: true })
})
```

### User Interactions
```typescript
const user = userEvent.setup()
const button = screen.getByRole('button', { name: 'Click Me' })
await user.click(button)
```

### Async Operations
```typescript
await waitFor(() => {
  expect(window.go.main.App.SomeFunction).toHaveBeenCalled()
})
```

## 📊 Best Practices

1. **Use semantic queries**: Prefer `getByRole`, `getByLabelText` over `getByTestId`
2. **Test user behavior**: Focus on what users see and do
3. **Mock external dependencies**: Mock Wails bindings and browser APIs
4. **Use async/await**: Handle asynchronous operations properly
5. **Clean up**: Clear mocks between tests
6. **Test error states**: Include error handling scenarios
7. **Accessibility**: Use accessible queries and test for a11y

## 🔍 Debugging Tests

### VS Code Integration
Add to `.vscode/settings.json`:
```json
{
  "vitest.enable": true,
  "vitest.commandLine": "npm run test"
}
```

### Debug in Browser
```bash
npm run test:ui
```

### Console Debugging
```typescript
import { screen } from '@testing-library/react'

// Debug what's rendered
screen.debug()

// Get all available queries
console.log(screen.logTestingPlaygroundURL())
```

## 📋 Coverage Goals

Aim for:
- **Lines**: >80%
- **Functions**: >80%
- **Branches**: >70%
- **Statements**: >80%

Critical paths should have 100% coverage:
- User authentication flows
- Data mutation operations
- Error handling
- Security-related functionality

## 🚨 Common Issues

### Mock Not Working
```typescript
// Ensure mocks are setup before component render
beforeEach(() => {
  vi.clearAllMocks()
  window.go.main.App.Function.mockResolvedValue(mockData)
})
```

### Async Test Failures
```typescript
// Always use waitFor for async operations
await waitFor(() => {
  expect(screen.getByText('Loading...')).not.toBeInTheDocument()
})
```

### Ant Design Components
```typescript
// Use the custom render function
import { render } from './test/test-utils' // Not from @testing-library/react
```

---

Happy testing! 🎉