import React, { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { ConfigProvider } from 'antd'
import ErrorBoundary from './ErrorBoundary'

// Custom render function that includes Ant Design ConfigProvider and Error Boundary
const AllTheProviders = ({ children }: { children: React.ReactNode }) => {
  return (
    <ErrorBoundary>
      <ConfigProvider>
        {children}
      </ConfigProvider>
    </ErrorBoundary>
  )
}

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>,
) => render(ui, { wrapper: AllTheProviders, ...options })

export * from '@testing-library/react'
export { customRender as render }