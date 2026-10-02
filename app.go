package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"oramcp/backend/audit"
	"oramcp/backend/integrations"
	"oramcp/backend/mcp"
	"oramcp/backend/models"
	"oramcp/backend/oracle"
	"oramcp/backend/security"
	"oramcp/backend/storage"
)

// App struct manages application state and exposes methods to Wails frontend.
type App struct {
	ctx       context.Context
	configMgr *storage.ConfigManager
	oracleMgr *oracle.ClientManager
	auditMgr  *audit.AuditManager
	mcpServer *mcp.MCPServer
}

// NewApp creates a new App application struct.
func NewApp() *App {
	cfg, err := storage.NewConfigManager()
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
	}

	ora := oracle.NewClientManager()
	aud := audit.NewAuditManager(500)
	srv := mcp.NewMCPServer(cfg, ora, aud)

	return &App{
		configMgr: cfg,
		oracleMgr: ora,
		auditMgr:  aud,
		mcpServer: srv,
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Hook audit listener to emit live frontend events
	a.auditMgr.AddListener(func(entry models.AuditLogEntry) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "mcp:audit", entry)
		}
	})

	// Start local MCP HTTP/SSE server
	port := a.configMgr.GetMCPPort()
	if err := a.mcpServer.StartHTTP(port); err != nil {
		fmt.Printf("Error starting MCP HTTP server: %v\n", err)
	} else {
		fmt.Printf("MCP HTTP/SSE server started on port %d\n", port)
	}
}

// domReady applies the initial window state once the native window is ready.
func (a *App) domReady(ctx context.Context) {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil {
		runtime.LogWarningf(ctx, "Could not detect startup screen: %v", err)
		return
	}
	for _, screen := range screens {
		if !screen.IsCurrent {
			continue
		}
		// Physical pixels identify 1080p even when OS display scaling is enabled.
		size := screen.PhysicalSize
		if size.Width <= 0 || size.Height <= 0 {
			size = screen.Size
		}
		if size.Width == 1920 && size.Height == 1080 {
			runtime.WindowMaximise(ctx)
		}
		return
	}
}

// shutdown is called when the app closes.
func (a *App) shutdown(ctx context.Context) {
	if a.mcpServer != nil {
		_ = a.mcpServer.StopHTTP()
	}
}

// --- Connection Management ---

func (a *App) GetConnections() []models.ConnectionProfile {
	return a.configMgr.GetConnections()
}

func (a *App) SaveConnection(profile models.ConnectionProfile) (*models.ConnectionProfile, error) {
	if profile.ID == "" {
		profile.ID = fmt.Sprintf("conn-%d", time.Now().UnixNano())
		profile.CreatedAt = time.Now()
	}
	profile.UpdatedAt = time.Now()

	// Invalidate any existing pool if host/user changed
	a.oracleMgr.InvalidatePool(profile.ID)

	if err := a.configMgr.SaveConnection(profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (a *App) DeleteConnection(id string) error {
	a.oracleMgr.InvalidatePool(id)
	return a.configMgr.DeleteConnection(id)
}

func (a *App) SetActiveConnection(id string) error {
	return a.configMgr.SetActiveConnection(id)
}

func (a *App) GetActiveConnection() *models.ConnectionProfile {
	return a.configMgr.GetActiveConnection()
}

func (a *App) TestConnection(profile models.ConnectionProfile) models.ConnectionTestResult {
	if profile.Password == "" && profile.ID != "" {
		if existing := a.configMgr.GetConnectionWithSecret(profile.ID); existing != nil && existing.Password != "" {
			profile.Password = existing.Password
		}
	}
	return a.oracleMgr.TestConnection(profile)
}

// --- Database Operations ---

func (a *App) resolveConnection(connID string) (*models.ConnectionProfile, error) {
	if connID != "" {
		conn := a.configMgr.GetConnectionWithSecret(connID)
		if conn != nil {
			return conn, nil
		}
	}
	active := a.configMgr.GetActiveConnection()
	if active == nil {
		return nil, fmt.Errorf("no Oracle connection available. Please configure and select a connection.")
	}
	return active, nil
}

func (a *App) GetSchemaObjects(connID string) (*models.SchemaInfo, error) {
	conn, err := a.resolveConnection(connID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.oracleMgr.GetSchemaObjects(ctx, *conn)
}

func (a *App) GetTableSchema(connID string, tableName string) ([]models.ColumnInfo, error) {
	conn, err := a.resolveConnection(connID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.oracleMgr.GetTableSchema(ctx, *conn, tableName)
}

func (a *App) ExecuteQuery(connID string, query string, maxRows int) (*models.QueryResult, error) {
	conn, err := a.resolveConnection(connID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return a.oracleMgr.ExecuteQuery(ctx, *conn, query, maxRows)
}

func (a *App) ExecutePLSQL(connID string, block string) (*models.PLSQLResult, error) {
	conn, err := a.resolveConnection(connID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return a.oracleMgr.ExecutePLSQL(ctx, *conn, block)
}

// --- Security & Policy Management ---

func (a *App) GetSecurityPolicy() models.SecurityPolicy {
	return a.configMgr.GetSecurityPolicy()
}

func (a *App) SaveSecurityPolicy(policy models.SecurityPolicy) error {
	return a.configMgr.SaveSecurityPolicy(policy)
}

func (a *App) ValidateQueryTest(query string) map[string]interface{} {
	policy := a.configMgr.GetSecurityPolicy()
	allowed, reason := security.ValidateQuery(query, policy)
	return map[string]interface{}{
		"allowed": allowed,
		"reason":  reason,
	}
}

// --- MCP Server & Audit Management ---

func (a *App) GetMCPServerStatus() models.MCPServerStatus {
	return a.mcpServer.GetStatus()
}

func (a *App) RestartMCPServer(port int) error {
	_ = a.mcpServer.StopHTTP()
	if port <= 0 {
		port = a.configMgr.GetMCPPort()
	}
	_ = a.configMgr.SetMCPPort(port)
	return a.mcpServer.StartHTTP(port)
}

func (a *App) GetAuditLogs() []models.AuditLogEntry {
	return a.auditMgr.GetEntries()
}

func (a *App) ClearAuditLogs() error {
	a.auditMgr.Clear()
	return nil
}

// --- Assisted & Automated Client Integrations ---

func (a *App) AutoConfigureClaude() (string, error) {
	port := a.configMgr.GetMCPPort()
	return integrations.AutoConfigureClaude(port)
}

func (a *App) AutoConfigureVSCode(targetDir string, transport string) (string, error) {
	port := a.configMgr.GetMCPPort()
	return integrations.AutoConfigureVSCodeWorkspace(targetDir, transport, port)
}

func (a *App) AutoConfigureCursor() (string, error) {
	port := a.configMgr.GetMCPPort()
	return integrations.AutoConfigureCursor(port)
}

// --- TNS Discovery & Parsing ---

func (a *App) SelectTNSFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context is not ready")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select tnsnames.ora",
		Filters: []runtime.FileFilter{
			{DisplayName: "Oracle TNS Files (*.ora)", Pattern: "*.ora"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
}

func (a *App) DetectTNSFiles() []string {
	return oracle.DetectTNSFiles()
}

func (a *App) ParseTNSFile(filePath string) ([]models.TNSEntry, error) {
	return oracle.ParseTNSFile(filePath)
}

// --- DBeaver Discovery & Parsing ---

func (a *App) SelectDBeaverFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context is not ready")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select DBeaver data-sources.json",
		Filters: []runtime.FileFilter{
			{DisplayName: "DBeaver Data Sources (data-sources.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
}

func (a *App) DetectDBeaverFiles() []string {
	return oracle.DetectDBeaverFiles()
}

func (a *App) ParseDBeaverFile(filePath string) ([]models.ConnectionProfile, error) {
	return oracle.ParseDBeaverFile(filePath)
}

// --- SQL File Management ---

func (a *App) OpenSQLFile() (map[string]string, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("application context not ready")
	}

	selected, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open SQL Script",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQL Scripts (*.sql)", Pattern: "*.sql"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || selected == "" {
		return nil, err
	}

	data, err := os.ReadFile(selected)
	if err != nil {
		return nil, fmt.Errorf("failed to read SQL file %s: %w", selected, err)
	}

	return map[string]string{
		"path":    selected,
		"name":    filepath.Base(selected),
		"content": string(data),
	}, nil
}

func (a *App) SaveSQLFile(filePath string, content string) (map[string]string, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("application context not ready")
	}

	if filePath == "" {
		selected, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title:           "Save SQL Script",
			DefaultFilename: "query.sql",
			Filters: []runtime.FileFilter{
				{DisplayName: "SQL Scripts (*.sql)", Pattern: "*.sql"},
				{DisplayName: "All Files (*.*)", Pattern: "*.*"},
			},
		})
		if err != nil || selected == "" {
			return nil, err
		}
		filePath = selected
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save SQL file %s: %w", filePath, err)
	}

	return map[string]string{
		"path": filePath,
		"name": filepath.Base(filePath),
	}, nil
}
