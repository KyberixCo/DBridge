package integrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	}

	// Linux
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "Claude", "claude_desktop_config.json"), nil
}

// GetVSCodeConfigPath returns the canonical path to VS Code mcp.json.
// If targetDir is specified, it targets the workspace (.vscode/mcp.json).
// If targetDir is empty, it targets the global User configuration (%APPDATA%\Code\User\mcp.json).
func GetVSCodeConfigPath(targetDir string) (string, error) {
	if targetDir != "" {
		return filepath.Join(targetDir, ".vscode", "mcp.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "mcp.json"), nil
	}

	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "mcp.json"), nil
	}

	// Linux
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "Code", "User", "mcp.json"), nil
}

// GetCursorConfigPath returns the canonical path to ~/.cursor/mcp.json.
func GetCursorConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor", "mcp.json"), nil
}

var reTrailingCommas = regexp.MustCompile(`,\s*([}\]])`)

// stripJSONCommentsAndTrailingCommas strips single/multiline comments and trailing commas from JSONC.
func stripJSONCommentsAndTrailingCommas(src []byte) []byte {
	var dst []byte
	inString := false
	escaped := false
	n := len(src)

	for i := 0; i < n; i++ {
		b := src[i]

		if inString {
			dst = append(dst, b)
			if escaped {
				escaped = false
			} else if b == '\\' {
				escaped = true
			} else if b == '"' {
				inString = false
			}
			continue
		}

		if b == '"' {
			inString = true
			dst = append(dst, b)
			continue
		}

		// Single-line comment //
		if b == '/' && i+1 < n && src[i+1] == '/' {
			for i < n && src[i] != '\n' && src[i] != '\r' {
				i++
			}
			if i < n {
				dst = append(dst, src[i])
			}
			continue
		}

		// Multi-line comment /* ... */
		if b == '/' && i+1 < n && src[i+1] == '*' {
			i += 2
			for i+1 < n && !(src[i] == '*' && src[i+1] == '/') {
				i++
			}
			i++
			continue
		}

		dst = append(dst, b)
	}

	// Remove trailing commas before closing braces/brackets
	return reTrailingCommas.ReplaceAll(dst, []byte("$1"))
}

// safeMergeJSONFile safely reads, modifies, and writes a JSON config file without overwriting existing data.
func safeMergeJSONFile(filePath string, mutator func(root map[string]interface{}) error) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	root := make(map[string]interface{})

	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		cleaned := stripJSONCommentsAndTrailingCommas(data)
		if unmarshalErr := json.Unmarshal(cleaned, &root); unmarshalErr != nil {
			// Do NOT overwrite user's file if unmarshaling fails!
			return fmt.Errorf("the file %s contains invalid JSON or syntax error; operation aborted to protect your existing configuration: %w", filePath, unmarshalErr)
		}
	}

	if err := mutator(root); err != nil {
		return err
	}

	newData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("error formatting JSON: %w", err)
	}

	if err := os.WriteFile(filePath, newData, 0644); err != nil {
		return fmt.Errorf("error writing to %s: %w", filePath, err)
	}

	return nil
}

// AutoConfigureClaude injects or updates the Oracle MCP server in Claude Desktop config.
func AutoConfigureClaude(port int) (string, error) {
	cfgPath, err := GetClaudeConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user directory: %w", err)
	}

	sseURL := fmt.Sprintf("http://localhost:%d/sse", port)

	err = safeMergeJSONFile(cfgPath, func(root map[string]interface{}) error {
		mcpServers, ok := root["mcpServers"].(map[string]interface{})
		if !ok || mcpServers == nil {
			mcpServers = make(map[string]interface{})
		}

		mcpServers["dbridge-oracle"] = map[string]interface{}{
			"url": sseURL,
		}
		root["mcpServers"] = mcpServers
		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Configuration successfully saved to %s", cfgPath), nil
}

// AutoConfigureVSCodeWorkspace creates or updates VS Code mcp.json.
// If targetDir is "", it updates the user's global configuration (%APPDATA%\Code\User\mcp.json on Windows).
func AutoConfigureVSCodeWorkspace(targetDir string, port int) (string, error) {
	mcpPath, err := GetVSCodeConfigPath(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve VS Code configuration path: %w", err)
	}

	sseURL := fmt.Sprintf("http://localhost:%d/sse", port)

	err = safeMergeJSONFile(mcpPath, func(root map[string]interface{}) error {
		// VS Code uses "servers" by default, or "mcpServers" depending on installed MCP extensions
		var targetMap map[string]interface{}
		targetKey := "servers"

		if servers, ok := root["servers"].(map[string]interface{}); ok && servers != nil {
			targetMap = servers
			targetKey = "servers"
		} else if mcpServers, ok := root["mcpServers"].(map[string]interface{}); ok && mcpServers != nil {
			targetMap = mcpServers
			targetKey = "mcpServers"
		} else {
			targetMap = make(map[string]interface{})
			targetKey = "servers"
		}

		// Update or insert dbridge-oracle while preserving all other configured servers
		targetMap["dbridge-oracle"] = map[string]interface{}{
			"type": "sse",
			"url":  sseURL,
		}

		root[targetKey] = targetMap
		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("VS Code configuration successfully updated at %s", mcpPath), nil
}

// AutoConfigureCursor injects or updates ~/.cursor/mcp.json
func AutoConfigureCursor(port int) (string, error) {
	mcpPath, err := GetCursorConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to resolve Cursor directory: %w", err)
	}

	sseURL := fmt.Sprintf("http://localhost:%d/sse", port)

	err = safeMergeJSONFile(mcpPath, func(root map[string]interface{}) error {
		mcpServers, ok := root["mcpServers"].(map[string]interface{})
		if !ok || mcpServers == nil {
			mcpServers = make(map[string]interface{})
		}

		mcpServers["dbridge-oracle"] = map[string]interface{}{
			"type": "sse",
			"url":  sseURL,
		}
		root["mcpServers"] = mcpServers
		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Cursor configuration successfully updated at %s", mcpPath), nil
}
