package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"oramcp/backend/audit"
	"oramcp/backend/models"
	"oramcp/backend/oracle"
	"oramcp/backend/storage"
)

func TestMCPServer_ProcessJSONRPC(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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

func TestToolReadOnlyAnnotations(t *testing.T) {
	server := NewMCPServer(nil, nil, nil)
	wantReadOnly := map[string]bool{
		"oracle_list_tables":      true,
		"oracle_describe_table":   true,
		"oracle_list_connections": true,
	}
	for _, tool := range server.getToolsList() {
		encoded, err := json.Marshal(tool)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		annotation, present := decoded["annotations"]
		if wantReadOnly[tool.Name] {
			if !present || annotation.(map[string]interface{})["readOnlyHint"] != true {
				t.Errorf("%s must declare readOnlyHint", tool.Name)
			}
		} else if present && annotation.(map[string]interface{})["readOnlyHint"] == true {
			t.Errorf("%s must not claim to be read-only", tool.Name)
		}
	}
}

func TestResourceURIRequiresAdvertisement(t *testing.T) {
	resources := []map[string]interface{}{{"uri": "oracle://USER/table/EMP", "name": "EMP"}}
	if got, ok := findResourceTable(resources, "oracle://USER/table/EMP"); !ok || got != "EMP" {
		t.Fatalf("advertised URI should resolve to EMP, got %q, %t", got, ok)
	}
	for _, uri := range []string{"oracle://USER/table/DEPT", "oracle://OTHER/table/EMP", "oracle://USER/table/EMP/../DEPT"} {
		if _, ok := findResourceTable(resources, uri); ok {
			t.Errorf("unadvertised URI %q was accepted", uri)
		}
	}
}

func TestResourceReadRejectsInvalidParams(t *testing.T) {
	server := NewMCPServer(nil, nil, nil)
	for _, params := range []string{`{}`, `{"uri":42}`, `{"uri":""}`} {
		request := `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":` + params + `}`
		response := server.ProcessJSONRPC(context.Background(), []byte(request), "test")
		var parsed struct {
			Error *RPCError `json:"error"`
		}
		if err := json.Unmarshal(response, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.Error == nil || parsed.Error.Code != -32602 {
			t.Errorf("params %s: expected -32602, got %s", params, response)
		}
	}
}

func TestInitializeNegotiatesProtocolVersion(t *testing.T) {
	server := NewMCPServer(nil, nil, nil)
	for _, test := range []struct {
		requested string
		want      string
	}{
		{"2024-11-05", "2024-11-05"},
		{"2025-03-26", "2025-03-26"},
		{"2099-01-01", "2025-03-26"},
	} {
		request := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"` + test.requested + `"}}`
		response := server.ProcessJSONRPC(context.Background(), []byte(request), "test")
		var parsed struct {
			Result struct {
				ProtocolVersion string `json:"protocolVersion"`
				Instructions    string `json:"instructions"`
			} `json:"result"`
		}
		if err := json.Unmarshal(response, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.Result.ProtocolVersion != test.want {
			t.Errorf("requested %s: want %s, got %s", test.requested, test.want, parsed.Result.ProtocolVersion)
		}
		if !strings.Contains(parsed.Result.Instructions, "MCP") {
			t.Errorf("initialize did not include MCP guidance")
		}
	}
}

func TestTableResourceContentsContainsSchemaOnly(t *testing.T) {
	columns := []models.ColumnInfo{{Name: "ID", DataType: "NUMBER", IsPrimaryKey: true}}
	contents, err := tableResourceContents("oracle://USER/table/EMP", "EMP", columns)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) != 1 || contents[0]["uri"] != "oracle://USER/table/EMP" || contents[0]["mimeType"] != "application/json" {
		t.Fatalf("invalid MCP resource content: %#v", contents)
	}
	var metadata struct {
		Table   string              `json:"table"`
		Columns []models.ColumnInfo `json:"columns"`
	}
	if err := json.Unmarshal([]byte(contents[0]["text"].(string)), &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.Table != "EMP" || len(metadata.Columns) != 1 || !metadata.Columns[0].IsPrimaryKey {
		t.Fatalf("resource did not return expected schema: %#v", metadata)
	}
}
