# Yak GUI Installation Guide

This guide provides multiple ways to install and run Yak GUI for your colleagues.

## Prerequisites

**Required:** The [yak CLI tool](https://github.com/santi1s/yak) must be installed and available in your PATH.

### Access Requirements
- **GitHub Access:** You need to be added as a collaborator to the private repositories.

### Install yak CLI
```bash
# Install yak CLI via Homebrew (private tap)
brew tap santi1s/tools
brew install yak
```

Verify installation:
```bash
yak --version
```

## Installation Options

### Option 1: Homebrew Installation (Recommended - Easiest)

```bash
# Add the tap (if not already added)
brew tap santi1s/tools

# Install yak-gui
brew install yak-gui
```

**Note:** This option requires the app to be published to the homebrew tap. If not available, use Option 2 or 3.

### Option 2: Download Pre-built Binary

1. Go to the [releases page](https://github.com/santi1s/yak-gui/releases)
2. Download the latest release for your platform:
   - **macOS:** `yak-gui-darwin-amd64.tar.gz` (Intel) or `yak-gui-darwin-arm64.tar.gz` (Apple Silicon)
   - **Linux:** `yak-gui-linux-amd64.tar.gz`
3. Extract the archive:
   ```bash
   tar -xzf yak-gui-*.tar.gz
   ```
4. Move to Applications (macOS) or install location (Linux):
   ```bash
   # macOS
   mv yak-gui.app /Applications/
   
   # Linux
   sudo mv yak-gui /usr/local/bin/
   ```

### Option 3: Build from Source

This option gives you the latest features and allows for customization.

#### Prerequisites
- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **Node.js 18+** - [Download here](https://nodejs.org/) or `brew install node`
- **Git** - Usually pre-installed or `brew install git`

#### Step-by-Step Instructions

1. **Clone the repository:**
   ```bash
   git clone https://github.com/santi1s/yak-gui.git
   cd yak-gui
   ```

2. **Install Wails CLI:**
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

3. **Install frontend dependencies:**
   ```bash
   cd frontend
   npm install
   cd ..
   ```

4. **Build the application:**
   ```bash
   wails build
   ```

5. **Run the application:**
   - **macOS:** Open `build/bin/yak-gui.app`
   - **Linux:** Run `./build/bin/yak-gui`

#### Development Mode (Optional)
For development with hot reload:
```bash
wails dev
```

## Running the Application

### macOS
- **From Applications:** Double-click the app icon
- **From Terminal:** `open /Applications/yak-gui.app`

### Linux
- **From Terminal:** `yak-gui` (if installed in PATH)
- **Direct execution:** `./yak-gui`

## Configuration

### Environment Setup
The app works best when launched from Terminal to inherit your shell environment. If launching from Finder/desktop:

1. Open the app
2. Go to the **Environment** tab
3. Click **"Import Shell Environment"** to load your AWS profiles, KUBECONFIG, etc.

### Required Environment Variables
- `AWS_PROFILE` - Your AWS profile
- `KUBECONFIG` - Path to your Kubernetes config
- `TFINFRA_REPOSITORY_PATH` - Path to terraform-infra repository (optional)

## Features Overview

### ArgoCD Management
- View applications in card or list view
- Sync, refresh, suspend/resume operations
- Real-time operation feedback
- Alphabetical sorting

### Argo Rollouts
- Monitor rollout status and history
- Promote, abort, restart, and retry operations
- Enhanced image display with SHA256 support

### Secrets Management
- Browse secrets hierarchically
- Create, update, and delete secrets
- JWT client/server configuration wizards

### Certificate Management
- SSL certificate renewal
- Certificate secret management
- Integration with Gandi DNS

### TFE Integration
- Workspace management
- Plan execution and monitoring
- Run management and controls

## Troubleshooting

### Common Issues

1. **"yak command not found"**
   ```bash
   # Check if yak is installed
   which yak
   
   # If not found, install it
   brew tap santi1s/tools
   brew install yak
   ```

2. **GUI doesn't see environment variables**
   - Launch from Terminal: `open /Applications/yak-gui.app`
   - Or use "Import Shell Environment" button in the Environment tab

3. **ArgoCD server not found**
   - Set your AWS_PROFILE: `export AWS_PROFILE=your-profile`
   - Ensure KUBECONFIG is set correctly
   - Use the Environment tab to configure profiles

4. **Build fails**
   ```bash
   # Check Go version (needs 1.21+)
   go version
   
   # Check Node.js version (needs 18+)
   node --version
   
   # Clear and reinstall dependencies
   rm -rf frontend/node_modules
   cd frontend && npm install
   ```

5. **App won't start**
   - Check system requirements
   - Try running from Terminal to see error messages
   - Ensure all prerequisites are installed

### Getting Debug Information

If you encounter issues:

1. **Run from Terminal** to see error messages:
   ```bash
   # macOS
   /Applications/yak-gui.app/Contents/MacOS/yak-gui
   
   # Linux
   ./yak-gui
   ```

2. **Check logs** in the Environment tab for shell import issues

3. **Verify yak CLI works** independently:
   ```bash
   yak argocd apps
   yak rollouts list
   ```

## System Requirements

- **macOS:** 10.13 (High Sierra) or later
- **Linux:** Ubuntu 18.04+ or equivalent (glibc 2.27+)
- **Memory:** 256MB RAM minimum
- **Disk:** 100MB available space

## Supported Platforms

- macOS (Intel and Apple Silicon)
- Linux (amd64)
- Windows support planned for future release

## Getting Help

- **Documentation:** Check the [README](README.md) for feature details
- **Issues:** Report bugs or request features on [GitHub Issues](https://github.com/santi1s/yak-gui/issues)
- **Updates:** Watch the repository for new releases
- **CLI Help:** Ensure you have the latest yak CLI: `brew upgrade yak`

## Security Notes

- The application inherits your shell environment and credentials
- All operations use the same credentials as your yak CLI
- No credentials are stored by the GUI application
- Use appropriate AWS profiles for different environments

## Development

If you want to contribute or modify the application:

1. Fork the repository
2. Follow the "Build from Source" instructions
3. Make your changes
4. Run tests: `cd frontend && npm test`
5. Submit a pull request

For more development details, see the [README](README.md).