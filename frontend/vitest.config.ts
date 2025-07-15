/// <reference types="vitest" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: true,
    silent: false,
    reporters: ['verbose'],
    testTimeout: 10000, // 10 seconds for slower tests
    hookTimeout: 10000, // 10 seconds for setup/teardown
    coverage: {
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        'dist/',
        'build/',
        'wailsjs/'
      ],
    },
    // Suppress console warnings from Ant Design in tests
    onConsoleLog(log, type) {
      if (type === 'stderr' && (
        log.includes('matchMedia') ||
        log.includes('Warning: An update to') ||
        log.includes('Warning: [antd: Menu]')
      )) {
        return false
      }
      return true
    },
  },
})