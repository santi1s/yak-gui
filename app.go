package main

import (
	"context"
	"fmt"
	
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct - Wails app context
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts, before the frontend is loaded
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Auto-import shell environment when launched from Finder
	// This helps ensure environment variables like GANDI_TOKEN are available
	if err := a.ImportShellEnvironment(); err != nil {
		fmt.Printf("Warning: failed to import shell environment on startup: %v\n", err)
	}
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Set window title
	runtime.WindowSetTitle(ctx, "Yak GUI")
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// Return true to prevent the application from quitting
	return false
}

// shutdown is called during application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform any teardown of resources here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// TestSimpleArray returns a simple array to test Wails binding
func (a *App) TestSimpleArray() []string {
	return []string{"app1", "app2", "app3"}
}

// GetAppVersion returns the application version information
func (a *App) GetAppVersion() map[string]string {
	return map[string]string{
		// x-release-please-start-version
		"version": "1.6.0",
		// x-release-please-end
		"name":    "Yak GUI",
		"description": "Comprehensive desktop GUI for yak CLI tool - manage ArgoCD, secrets, rollouts, certificates, and more",
	}
}

// MaximizeWindow maximizes the application window
func (a *App) MaximizeWindow() {
	if a.ctx != nil {
		runtime.WindowMaximise(a.ctx)
	}
}

// UnmaximizeWindow restores the application window from maximized state
func (a *App) UnmaximizeWindow() {
	if a.ctx != nil {
		runtime.WindowUnmaximise(a.ctx)
	}
}

// IsWindowMaximized returns whether the window is currently maximized
func (a *App) IsWindowMaximized() bool {
	if a.ctx != nil {
		return runtime.WindowIsMaximised(a.ctx)
	}
	return false
}