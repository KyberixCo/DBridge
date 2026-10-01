package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"oramcp/backend/audit"
	"oramcp/backend/mcp"
	"oramcp/backend/oracle"
	"oramcp/backend/storage"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check for headless stdio MCP server mode (e.g. `dbridge --mcp` or `dbridge stdio`)
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		if arg == "--mcp" || arg == "-mcp" || arg == "mcp" || arg == "stdio" || arg == "--stdio" || arg == "-stdio" || arg == "/mcp" {
			runStdioMCP()
			return
		}
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options matching Crimson-Duck-DM1 frameless architecture
	err := wails.Run(&options.App{
		Title:             "KYBERIX // DBRIDGE",
		Width:             1280,
		Height:            850,
		MinWidth:          960,
		MinHeight:         640,
		Frameless:         true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:  &options.RGBA{R: 5, G: 5, B: 5, A: 255}, // #050505
		OnStartup:         app.startup,
		OnShutdown:        app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			Theme:                             windows.Dark,
			DisableFramelessWindowDecorations: false,
		},
		Mac: &mac.Options{
			TitleBar:             nil, // Completely custom window controls rendered in frontend
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "KYBERIX // DBRIDGE",
				Message: "Tactical Oracle Database Model Context Protocol Bridge\nKyberix Corp.",
			},
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func runStdioMCP() {
	cfg, err := storage.NewConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dbridge: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	ora := oracle.NewClientManager()
	aud := audit.NewAuditManager(500)
	server := mcp.NewMCPServer(cfg, ora, aud)

	ctx := context.Background()
	if err := server.RunStdio(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "dbridge stdio exited with error: %v\n", err)
		os.Exit(1)
	}
}
