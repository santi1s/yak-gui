# Yak ArgoCD GUI

A modern desktop GUI for managing ArgoCD applications using Wails v2 and your existing Yak CLI infrastructure.

## Features

- **Real-time Dashboard**: View all ArgoCD applications with health and sync status
- **Application Management**: Sync, refresh, suspend/unsuspend applications
- **Status Monitoring**: Health status, sync status, and sync loop detection
- **Native Performance**: Built with Wails v2 for native desktop experience
- **Existing Integration**: Uses your existing Yak CLI ArgoCD commands

## Prerequisites

- Go 1.22+
- Node.js 18+
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Quick Start

1. **Install dependencies**:
   ```bash
   cd yak-gui
   go mod tidy
   cd frontend && npm install
   ```

2. **Run in development mode**:
   ```bash
   wails dev
   ```

3. **Build for production**:
   ```bash
   wails build
   ```

## Configuration

The GUI connects to your ArgoCD server using the same authentication methods as your CLI:

- **Server**: Your ArgoCD server URL (e.g., `argocd.example.com`)
- **Project**: ArgoCD project name (defaults to `main`)
- **Authentication**: Uses OIDC authentication (same as CLI)

## Architecture

### Backend (`app.go`)
- Wraps your existing ArgoCD CLI functions
- Uses same authentication and data structures
- Provides Go API for frontend consumption

### Frontend (React + TypeScript)
- Modern React dashboard with Tailwind CSS
- Real-time application monitoring
- One-click actions for common operations

## Available Operations

- **View Applications**: List all applications with status indicators
- **Sync Applications**: Trigger synchronization (with prune/dry-run options)
- **Refresh Applications**: Force refresh application state
- **Suspend/Unsuspend**: Control application sync windows
- **Auto-refresh**: Automatic dashboard updates every 30 seconds

## Status Indicators

### Health Status
- 🟢 **Healthy**: Application is running correctly
- 🟡 **Progressing**: Application is being deployed
- 🔴 **Degraded**: Application has issues
- ⚫ **Suspended**: Application sync is suspended
- 🟠 **Missing**: Resources are missing

### Sync Status
- 🟢 **Synced**: Application is in sync with Git
- 🔴 **OutOfSync**: Application differs from Git

### Sync Loop Detection
- 🔴 **Critical**: 3+ syncs in 15 minutes
- 🟡 **Warning**: 6+ syncs in 1 hour
- 🟠 **Possible**: Currently syncing with recent activity

## Development

The project structure follows Wails v2 conventions:

```
yak-gui/
├── app.go              # Go backend API
├── main.go             # Wails application entry
├── go.mod              # Go dependencies
├── wails.json          # Wails configuration
└── frontend/           # React frontend
    ├── src/
    │   ├── App.tsx     # Main React component
    │   ├── main.tsx    # React entry point
    │   └── style.css   # Tailwind styles
    ├── package.json    # Node dependencies
    └── vite.config.ts  # Vite configuration
```

## Integration with Yak CLI

This GUI reuses your existing Yak CLI code:

- **Authentication**: Same OIDC flow as `yak argocd` commands
- **API Calls**: Uses `argocdhelper` package directly
- **Data Structures**: Same `statusMap` and `ArgoApp` types
- **Error Handling**: Consistent error messaging

## Building for Distribution

```bash
# Build for current platform
wails build

# Build for all platforms
wails build -platform darwin/amd64,darwin/arm64,linux/amd64,windows/amd64
```

The built application will be in the `build/bin/` directory.

## Troubleshooting

### Common Issues

1. **Go module not found**: Make sure the `replace` directive in `go.mod` points to your Yak CLI directory
2. **ArgoCD connection failed**: Verify your ArgoCD server URL and authentication
3. **Frontend build fails**: Run `npm install` in the `frontend/` directory

### Logs

Run with debug logging:
```bash
wails dev -loglevel Debug
```