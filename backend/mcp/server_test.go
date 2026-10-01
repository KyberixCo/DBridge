package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"oramcp/backend/audit"
	"oramcp/backend/oracle"
	"oramcp/backend/storage"
)

func TestMCPServer_ProcessJSONRPC(t *testing.T) {
	cfg, err := storage.NewConfigManager()
	if err != nil {
		t.Fatalf("Failed to init config manager: %v", err)
	}
	ora := oracle.NewClientManager()
	aud := audit.NewAuditManager(100)
	server := NewMCPServer(cfg, ora, aud)

	ctx := context.Background()

	// 1. Test initialize
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`
	resp := server.ProcessJSONRPC(ctx, []byte(initReq), "test")
	var r map[string]interface{}
	if err := json.Unmarshal(resp, &r); err != nil {
		t.Fatalf("Failed to unmarshal initialize response: %v", err)
	}
	if r["error"] != nil {
		t.Fatalf("Expected no error, got: %v", r["error"])
	}

	// 2. Test tools/list
	listReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	resp = server.ProcessJSONRPC(ctx, []byte(listReq), "test")
	if err := json.Unmarshal(resp, &r); err != nil {
		t.Fatalf("Failed to unmarshal tools/list response: %v", err)
	}
	result := r["result"].(map[string]interface{})
	tools := result["tools"].([]interface{})
	if len(tools) == 0 {
		t.Fatalf("Expected tools in tools/list")
	}

	// 3. Test security block on tools/call oracle_query with DELETE
	queryCall := `{
		"jsonrpc": "2.0",
		"id": 3,
		"method": "tools/call",
		"params": {
			"name": "oracle_query",
			"arguments": {
				"query": "DELETE FROM EMPLOYEES WHERE ID = 1"
			}
		}
	}`
	resp = server.ProcessJSONRPC(ctx, []byte(queryCall), "test")
	var callResp map[string]interface{}
	_ = json.Unmarshal(resp, &callResp)
	callResult := callResp["result"].(map[string]interface{})
	if callResult["isError"] != true {
		t.Errorf("Expected DELETE query to be flagged as error (blocked by policy)")
	}
	content := callResult["content"].([]interface{})
	firstContent := content[0].(map[string]interface{})
	text := firstContent["text"].(string)
	if !strings.Contains(text, "SECURITY BLOCKED") {
		t.Errorf("Expected text to mention SECURITY BLOCKED, got: %s", text)
	}

	// Verify it was logged in audit log
	entries := aud.GetEntries()
	if len(entries) == 0 {
		t.Errorf("Expected audit entry to be recorded")
	} else if entries[0].Allowed {
		t.Errorf("Audit entry should show Allowed=false")
	}

	// 4. Test prompts/list
	promptsReq := `{"jsonrpc":"2.0","id":4,"method":"prompts/list"}`
	resp = server.ProcessJSONRPC(ctx, []byte(promptsReq), "test")
	var promptsResp map[string]interface{}
	if err := json.Unmarshal(resp, &promptsResp); err != nil {
		t.Fatalf("Failed to unmarshal prompts/list response: %v", err)
	}
	if promptsResp["error"] != nil {
		t.Errorf("Expected prompts/list to succeed, got error: %v", promptsResp["error"])
	}

	// 5. Test resources/templates/list
	tmplReq := `{"jsonrpc":"2.0","id":5,"method":"resources/templates/list"}`
	resp = server.ProcessJSONRPC(ctx, []byte(tmplReq), "test")
	var tmplResp map[string]interface{}
	if err := json.Unmarshal(resp, &tmplResp); err != nil {
		t.Fatalf("Failed to unmarshal resources/templates/list response: %v", err)
	}
	if tmplResp["error"] != nil {
		t.Errorf("Expected resources/templates/list to succeed, got error: %v", tmplResp["error"])
	}

	// 6. Test initialized notification returns nil
	initNotif := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	if r := server.ProcessJSONRPC(ctx, []byte(initNotif), "test"); len(r) > 0 {
		t.Errorf("Expected notifications/initialized to produce no response, got: %s", string(r))
	}

	initNotif2 := `{"jsonrpc":"2.0","method":"initialized"}`
	if r := server.ProcessJSONRPC(ctx, []byte(initNotif2), "test"); len(r) > 0 {
		t.Errorf("Expected initialized to produce no response, got: %s", string(r))
	}

	// 7. Unknown notification without id returns nil per JSON-RPC 2.0
	unknownNotif := `{"jsonrpc":"2.0","method":"random/notification"}`
	if r := server.ProcessJSONRPC(ctx, []byte(unknownNotif), "test"); len(r) > 0 {
		t.Errorf("Expected unknown notification to produce no response, got: %s", string(r))
	}
}
