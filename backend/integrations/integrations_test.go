package integrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripJSONCommentsAndTrailingCommas(t *testing.T) {
	input := []byte(`{
		// This is a single line comment
		"servers": {
			"existing-server": {
				"url": "http://example.com/api//not-a-comment", /* inline comment */
				"type": "sse",
			},
		},
	}`)

	cleaned := stripJSONCommentsAndTrailingCommas(input)

	var parsed map[string]interface{}
	if err := json.Unmarshal(cleaned, &parsed); err != nil {
		t.Fatalf("failed to parse cleaned JSON: %v\nCleaned content was:\n%s", err, string(cleaned))
	}

	servers, ok := parsed["servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("servers key missing")
	}

	existing, ok := servers["existing-server"].(map[string]interface{})
	if !ok {
		t.Fatalf("existing-server missing")
	}

	if existing["url"] != "http://example.com/api//not-a-comment" {
		t.Errorf("url with // in string was corrupted: %v", existing["url"])
	}
}

func TestAutoConfigureVSCodeWorkspace_PreservesExisting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vscode-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vscodeDir := filepath.Join(tempDir, ".vscode")
	if err := os.MkdirAll(vscodeDir, 0755); err != nil {
		t.Fatalf("failed to create .vscode dir: %v", err)
	}

	initialConfig := `{
		// Existing VS Code MCP setup
		"servers": {
			"custom-server": {
				"type": "stdio",
				"command": "python",
				"args": ["server.py"],
			},
		},
	}`

	mcpPath := filepath.Join(vscodeDir, "mcp.json")
	if err := os.WriteFile(mcpPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	// Run configuration
	msg, err := AutoConfigureVSCodeWorkspace(tempDir, 9090)
	if err != nil {
		t.Fatalf("AutoConfigureVSCodeWorkspace failed: %v", err)
	}

	if !strings.Contains(msg, mcpPath) {
		t.Errorf("expected path in response message, got: %s", msg)
	}

	// Verify file contents
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	var updated map[string]interface{}
	if err := json.Unmarshal(data, &updated); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v\nData: %s", err, string(data))
	}

	servers, ok := updated["servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("servers key missing from updated config")
	}

	// Verify custom-server was preserved
	if _, exists := servers["custom-server"]; !exists {
		t.Errorf("existing server 'custom-server' was wiped out!")
	}

	// Verify dbridge-oracle was added
	dbridge, exists := servers["dbridge-oracle"].(map[string]interface{})
	if !exists {
		t.Fatalf("dbridge-oracle was not added to servers")
	}

	if dbridge["url"] != "http://localhost:9090/sse" {
		t.Errorf("expected URL http://localhost:9090/sse, got: %v", dbridge["url"])
	}
}
