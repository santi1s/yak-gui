package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

// App instance that will be exposed to frontend
var app *App

// main starts the Wails application
func main() {
	// Create an instance of the app structure
	app = NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "Yak ArgoCD GUI",
		Width:            1200,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		MaxWidth:         0,  // 0 means no limit
		MaxHeight:        0,  // 0 means no limit
		DisableResize:    false,
		Fullscreen:       false,
		StartHidden:      false,
		HideWindowOnClose: false,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:                 &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:                        app.startup,
		OnDomReady:                       app.domReady,
		OnBeforeClose:                    app.beforeClose,
		OnShutdown:                       app.shutdown,
		EnableDefaultContextMenu:         true,
		EnableFraudulentWebsiteDetection: false,
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "Yak ArgoCD GUI",
				Message: "Desktop GUI for yak CLI tool",
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}