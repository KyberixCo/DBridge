package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	if stdioModeIndex(os.Args[1:]) >= 0 {
		initConsole()
		if err := runStdioMCP(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "dbridge: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options matching Crimson-Duck-DM1 frameless architecture
	err := wails.Run(&options.App{
		Title:     "KYBERIX // DBRIDGE",
		Width:     1280,
		Height:    850,
		MinWidth:  960,
		MinHeight: 640,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 5, G: 5, B: 5, A: 255}, // #050505
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
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

func runStdioMCP(args []string) error {
	opts, err := parseStdioOptions(args)
	if err != nil {
		return err
	}
	cfg, err := storage.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	ora := oracle.NewClientManager()
	aud := audit.NewAuditManager(500)
	server := mcp.NewMCPServer(cfg, ora, aud)
	if err := server.SetRequestTimeout(opts.timeout); err != nil {
		return err
	}
	if opts.inspect {
		if err := server.ConfigureInspection(opts.connectionID); err != nil {
			return err
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.RunStdio(ctx); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("stdio exited with error: %w", err)
	}
	return nil
}
