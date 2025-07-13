# yak-gui Release Guide

This document describes how to build and release the Wails desktop GUI for the yak CLI tool using your existing CI/CD pipeline and Homebrew tap.

## 🍺 Integration with Existing Homebrew Tap

Since you already have a CI/CD pipeline distributing `yak` via Homebrew tap, you can extend it to also distribute the Wails GUI app.

## 🔨 CI/CD Pipeline Integration

### Add Wails Build Step to Your Existing Workflow

Add this to your GitHub Actions workflow (or equivalent CI system):

```yaml
# Example GitHub Actions workflow addition
- name: Build Wails GUI
  run: |
    cd yak-gui
    wails build --platform darwin/amd64,darwin/arm64
    # Create archives for distribution
    tar -czf yak-gui-darwin-amd64.tar.gz -C build/bin yak-gui.app
    tar -czf yak-gui-darwin-arm64.tar.gz -C build/bin yak-gui.app

- name: Build Universal Binary (Optional)
  run: |
    cd yak-gui
    wails build --platform darwin/universal
    tar -czf yak-gui-darwin-universal.tar.gz -C build/bin yak-gui.app
```

### Multi-Platform Build Matrix (Optional)

For broader platform support:

```yaml
strategy:
  matrix:
    include:
      - os: macos-latest
        platform: darwin/amd64,darwin/arm64
        archive_name: darwin
      - os: ubuntu-latest  
        platform: linux/amd64
        archive_name: linux
      - os: windows-latest
        platform: windows/amd64
        archive_name: windows

steps:
  - name: Build Wails App
    run: |
      cd yak-gui
      wails build --platform ${{ matrix.platform }}
```

## 🍺 Homebrew Formula

Add this formula to your existing Homebrew tap repository:

```ruby
# Formula/yak-gui.rb
class YakGui < Formula
  desc "Desktop GUI for yak CLI tool"
  homepage "https://github.com/your-org/yak"
  version "1.0.0"

  if Hardware::CPU.arm?
    url "https://github.com/your-org/yak/releases/download/v#{version}/yak-gui-darwin-arm64.tar.gz"
    sha256 "arm64_sha256_here"
  else
    url "https://github.com/your-org/yak/releases/download/v#{version}/yak-gui-darwin-amd64.tar.gz"
    sha256 "amd64_sha256_here"
  end

  depends_on "yak" # Ensure yak CLI is installed

  def install
    # Install the .app bundle
    prefix.install "yak-gui.app"
    
    # Create symlink in Applications for easy access
    system "ln", "-sf", "#{prefix}/yak-gui.app", "/Applications/yak-gui.app"
    
    # Optional: Create CLI wrapper for command-line access
    bin.write_exec_script "#{prefix}/yak-gui.app/Contents/MacOS/yak-gui"
  end

  def caveats
    <<~EOS
      yak-gui has been installed to:
        #{prefix}/yak-gui.app

      A symlink has been created in /Applications for easy access.
      
      You can also run it from the command line with: yak-gui
    EOS
  end

  test do
    # Test that the app bundle exists
    assert_predicate prefix/"yak-gui.app", :exist?
  end
end
```

## 🚀 Release Process

### Update Your Existing Release Workflow

Modify your current release process to include GUI artifacts:

```yaml
- name: Create Release
  uses: softprops/action-gh-release@v1
  with:
    files: |
      yak-*
      yak-gui-*
    tag_name: ${{ github.ref_name }}
    name: Release ${{ github.ref_name }}
    draft: false
    prerelease: false
```

### Version Synchronization

Keep yak-gui version in sync with yak CLI:
- Use the same version number for both
- Tag releases consistently
- Update both formulas simultaneously

## 📦 User Installation

Once integrated, users can install via your existing tap:

```bash
# Install your tap (if not already added)
brew tap your-org/tap

# Install yak CLI
brew install yak

# Install yak GUI
brew install yak-gui

# Or install both at once
brew install yak yak-gui

# Update both tools
brew upgrade yak yak-gui
```

## 🎯 Alternative: Unified Formula

You could also create a single formula that includes both CLI and GUI:

```ruby
class YakSuite < Formula
  desc "Complete yak toolkit with CLI and GUI"
  # ... includes both yak binary and yak-gui.app
  
  def install
    bin.install "yak"
    prefix.install "yak-gui.app"
    # ... 
  end
end
```

## 🔧 Build Requirements

Ensure your CI environment has:
- Go 1.19+ for yak CLI
- Node.js 16+ for Wails frontend
- Wails CLI installed: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## ✅ Testing the Release

Before pushing to production:

1. Test the build process locally:
   ```bash
   cd yak-gui
   wails build
   ```

2. Test the Homebrew formula:
   ```bash
   brew install --build-from-source ./Formula/yak-gui.rb
   ```

3. Verify the app launches and functions correctly

## 🌟 Advantages of This Approach

- **Unified Distribution**: Same tap for all your tools
- **Dependency Management**: GUI automatically depends on CLI  
- **Easy Updates**: `brew upgrade` updates both tools
- **Familiar Workflow**: Users already know how to use your tap
- **Cross-Platform Ready**: Can extend to Linux/Windows if needed
- **Consistent Versioning**: Keep CLI and GUI in sync

## 📋 Release Checklist

- [ ] Add Wails build steps to CI pipeline
- [ ] Create yak-gui.rb formula in your tap repository
- [ ] Update release workflow to include GUI artifacts
- [ ] Test installation process locally
- [ ] Update project documentation
- [ ] Create release notes mentioning GUI availability
- [ ] Announce new GUI option to users

## 🐛 Troubleshooting

### Common Issues

1. **Build Failures**: Ensure Wails CLI is installed in CI environment
2. **App Won't Launch**: Check that the yak binary is in PATH
3. **Permission Issues**: Verify app signing if distributing outside App Store
4. **Dependencies**: Ensure yak formula is available before yak-gui installation

### Debug Commands

```bash
# Check Wails installation
wails doctor

# Verify build output
ls -la build/bin/

# Test formula syntax
brew audit --strict Formula/yak-gui.rb
```