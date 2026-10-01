package integrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ClientInfo details integration targets.
type ClientInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // "auto" or "manual"
	ConfigPath  string `json:"configPath"`
	Description string `json:"description"`
}

// GetClaudeConfigPath returns the canonical path to claude_desktop_config.json.
func GetClaudeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	}

	// macOS
	return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
}

// AutoConfigureClaude injects or updates the Oracle MCP server in Claude Desktop config.
func AutoConfigureClaude(port int) (string, error) {
	cfgPath, err := GetClaudeConfigPath()
	if err != nil {
		return "", fmt.Errorf("no se pudo determinar el directorio de usuario: %w", err)
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error al crear el directorio de Claude: %w", err)
	}

	// Read existing config or initialize
	var root map[string]interface{}
	data, err := os.ReadFile(cfgPath)
	if err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]interface{})
	}

	mcpServers, ok := root["mcpServers"].(map[string]interface{})
	if !ok || mcpServers == nil {
		mcpServers = make(map[string]interface{})
	}

	sseURL := fmt.Sprintf("http://localhost:%d/sse", port)
	mcpServers["oracle"] = map[string]interface{}{
		"url": sseURL,
	}
	root["mcpServers"] = mcpServers

	newData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error formateando JSON: %w", err)
	}

	if err := os.WriteFile(cfgPath, newData, 0644); err != nil {
		return "", fmt.Errorf("error escribiendo archivo: %w", err)
	}

	return fmt.Sprintf("Configuración inyectada con éxito en %s", cfgPath), nil
}

// AutoConfigureVSCodeWorkspace creates or updates .vscode/mcp.json in the current/selected directory.
func AutoConfigureVSCodeWorkspace(targetDir string, port int) (string, error) {
	if targetDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		targetDir = cwd
	}

	vscodeDir := filepath.Join(targetDir, ".vscode")
	if err := os.MkdirAll(vscodeDir, 0755); err != nil {
		return "", fmt.Errorf("error creando carpeta .vscode: %w", err)
	}

	mcpPath := filepath.Join(vscodeDir, "mcp.json")
	var root map[string]interface{}
	data, err := os.ReadFile(mcpPath)
	if err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]interface{})
	}

	servers, ok := root["servers"].(map[string]interface{})
	if !ok || servers == nil {
		servers = make(map[string]interface{})
	}

	sseURL := fmt.Sprintf("http://localhost:%d/sse", port)
	servers["oracle"] = map[string]interface{}{
		"type": "sse",
		"url":  sseURL,
	}
	root["servers"] = servers

	newData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializando .vscode/mcp.json: %w", err)
	}

	if err := os.WriteFile(mcpPath, newData, 0644); err != nil {
		return "", fmt.Errorf("error escribiendo .vscode/mcp.json: %w", err)
	}

	return fmt.Sprintf("Archivo .vscode/mcp.json generado con éxito en %s", mcpPath), nil
}

// AutoConfigureCursor injects into ~/.cursor/mcp.json
func AutoConfigureCursor(port int) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		return "", err
	}

	mcpPath := filepath.Join(cursorDir, "mcp.json")
	var root map[string]interface{}
	data, err := os.ReadFile(mcpPath)
	if err == nil {
		_ = json.Unmarshal(data, &root)
	}
	if root == nil {
		root = make(map[string]interface{})
	}

	mcpServers, ok := root["mcpServers"].(map[string]interface{})
	if !ok || mcpServers == nil {
		mcpServers = make(map[string]interface{})
	}

	sseURL := fmt.Sprintf("http://localhost:%d/sse", port)
	mcpServers["oracle_bridge"] = map[string]interface{}{
		"type": "sse",
		"url":  sseURL,
	}
	root["mcpServers"] = mcpServers

	newData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(mcpPath, newData, 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("Configuración inyectada con éxito en %s", mcpPath), nil
}
