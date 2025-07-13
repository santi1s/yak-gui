# yak - CLI and GUI Tools

A comprehensive toolkit for managing ArgoCD applications, Argo Rollouts, and Vault secrets with both command-line and desktop GUI interfaces.

## 🚀 Features

### CLI Tool (`yak`)
- **ArgoCD Management**: Status, sync, suspend applications
- **Argo Rollouts**: Promote, pause, abort, restart rollouts  
- **Secrets Management**: Create, read, update, delete Vault secrets
- **JWT Tokens**: Generate JWT client/server secrets
- **Terraform**: Module and provider management
- **AWS**: Aurora, ECR, and configuration tools
- **Certificate Management**: SSL/TLS certificate operations

### Desktop GUI (`yak-gui`)
- **Modern Interface**: Built with Wails v2 (Go + React)
- **ArgoCD Dashboard**: Visual application management
- **Rollouts Console**: Interactive rollout operations with image tracking
- **Secrets Browser**: File explorer-style secret navigation
- **JWT Tools**: User-friendly JWT secret creation

## 📦 Installation

### Via Homebrew (Recommended)

```bash
# Add the tap
brew tap santi1s/tools https://github.com/santi1s/homebrew-tools

# Install CLI
brew install yak

# Install GUI
brew install yak-gui

# Or install both
brew install yak yak-gui
```

### Via Docker

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/santi1s/yak:latest

# Run
docker run --rm ghcr.io/santi1s/yak:latest --help
```

### From Source

```bash
# Requirements: Go 1.21+
git clone https://github.com/santi1s/yak.git
cd yak

# Build CLI
go build ./cmd/yak

# Build GUI (requires Wails v2 and Node.js 18+)
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd yak-gui
wails build
```

## 🖥️ GUI Features

### ArgoCD Tab
- Real-time application status monitoring
- One-click sync/suspend operations
- AWS profile auto-detection
- SAML authentication handling

### Rollouts Tab  
- Visual rollout cards with status indicators
- **Container image tracking** with version tags
- Promote, pause, abort, restart actions
- Image update functionality
- Strategy visualization (Canary, Blue-Green)

### Secrets Tab
- Hierarchical secret browsing with breadcrumbs
- Create, edit, delete operations
- **JWT client/server secret wizards**
- Path-based navigation
- File explorer-style interface

## 🔧 CLI Usage Examples

### ArgoCD
```bash
# Check application status
yak argocd status

# Sync an application
yak argocd sync my-app

# Monitor applications
yak argocd monitor
```

### Rollouts
```bash
# List rollouts
yak rollouts list

# Promote a rollout
yak rollouts promote my-rollout

# Check rollout status
yak rollouts status -r my-rollout
```

### Secrets
```bash
# List secrets
yak secret list --platform dev

# Get a secret
yak secret get --path myapp/database --platform dev

# Create JWT tokens
yak secret jwt client --help
yak secret jwt server --help
```

## 🏗️ Development

### Prerequisites
- Go 1.21+
- Node.js 18+ (for GUI)
- Wails v2 CLI (for GUI)

### Building
```bash
# CLI
go build ./cmd/yak

# GUI
cd yak-gui
wails build

# Development mode (GUI)
wails dev
```

### Testing
```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/cmd/argocd

# Linting
golangci-lint run ./...
```

## 🚀 Release Process

This project uses GitHub Actions for automated releases:

1. **Create a tag**: `git tag v1.0.0 && git push origin v1.0.0`
2. **GitHub Actions automatically**:
   - Builds CLI for multiple platforms
   - Builds Wails GUI for macOS
   - Creates Docker image and pushes to GHCR
   - Creates GitHub release with binaries
3. **Update Homebrew formulas** with new SHA256 checksums

See [yak-gui-release.md](yak-gui/yak-gui-release.md) for detailed release instructions.

## 📋 Configuration

The GUI automatically inherits configuration from the yak CLI. Ensure your CLI is properly configured before using the GUI.

### Environment Variables
- `AWS_PROFILE`: AWS profile for ArgoCD server detection
- `KUBECONFIG`: Kubernetes configuration for rollouts

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests if applicable
5. Run lints: `golangci-lint run ./...`
6. Commit your changes: `git commit -m 'feat: add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Submit a pull request

## 📄 License

This project is private and proprietary.

## 🔗 Links

- **Repository**: https://github.com/santi1s/yak
- **Homebrew Tap**: https://github.com/santi1s/homebrew-tools
- **Container Registry**: ghcr.io/santi1s/yak
- **Release Guide**: [yak-gui-release.md](yak-gui/yak-gui-release.md)