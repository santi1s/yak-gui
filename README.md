# Yak GUI

A desktop GUI application for the [yak CLI tool](https://github.com/santi1s/yak), built with [Wails](https://wails.io/).

## Features

- **ArgoCD Management**: View and manage ArgoCD applications with intuitive interface
- **Argo Rollouts**: Monitor and control rollout deployments with enhanced image display
- **Secrets Management**: Browse and manage secrets with hierarchical navigation
- **JWT Tools**: Create JWT client/server configurations with guided wizards
- **Cross-Platform**: Available for macOS and Linux

## Installation

### macOS (Recommended)

Download the latest release from the [releases page](https://github.com/santi1s/yak-gui/releases) and extract the `.tar.gz` file.

### Prerequisites

The GUI application requires the [yak CLI tool](https://github.com/santi1s/yak) to be installed and available in your PATH.

```bash
# Install yak CLI via Homebrew
brew tap santi1s/tools
brew install yak
```

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- Wails v2

### Setup

```bash
# Clone the repository
git clone https://github.com/santi1s/yak-gui.git
cd yak-gui

# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run in development mode
wails dev

# Build for production
wails build
```

### Architecture

- **Backend**: Go with Wails v2 framework
- **Frontend**: React with TypeScript and Tailwind CSS
- **Build**: Cross-platform builds via GitHub Actions

## Features Detail

### Enhanced Rollout Display
- Support for both Docker tags and SHA256 digests
- Truncated SHA256 display with full digest on hover
- Color-coded tags (purple) vs digests (orange)
- Responsive layout handling long image names

### macOS PATH Resolution
- Automatic discovery of yak CLI in common installation paths
- Handles GUI applications' limited PATH environment
- Fallback to standard PATH resolution

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `npm test` (frontend) and `wails build` (full app)
5. Submit a pull request

## License

This project follows the same license as the main yak CLI tool.