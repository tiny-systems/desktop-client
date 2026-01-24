package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	ctrlog "sigs.k8s.io/controller-runtime/pkg/log"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	l := NewWailsLogr(logger.NewDefaultLogger())
	// Create an instance of the app structure
	app := NewApp(l)

	ctrlog.SetLogger(l)

	// Create application menu (required for Cmd shortcuts on macOS)
	var appMenu *menu.Menu
	if runtime.GOOS == "darwin" {
		appMenu = menu.NewMenuFromItems(
			menu.AppMenu(),
			menu.EditMenu(), // Enables Cmd+C/V/X/A and other Cmd key events
		)
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TinySystems",
		Width:  1280,
		Height: 1024,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Menu: appMenu, // Enable Cmd shortcuts on macOS

		Logger:             nil, // Uses default logger
		LogLevel:           logger.INFO,
		LogLevelProduction: logger.ERROR,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
