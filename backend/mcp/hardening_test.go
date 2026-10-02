package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/zalando/go-keyring"

	"oramcp/backend/audit"
	"oramcp/backend/models"
	"oramcp/backend/storage"
)

type testOracle struct {
	started      chan struct{}
	finished     chan struct{}
	queryRows    int
	queryProfile string
	readOnly     bool
}

func (o *testOracle) GetSchemaObjects(ctx context.Context, _ models.ConnectionProfile) (*models.SchemaInfo, error) {
	if o.started != nil {
		close(o.started)
		<-ctx.Done()
		close(o.finished)
		return nil, ctx.Err()
	}
	return &models.SchemaInfo{Tables: []string{"EMP"}}, nil
}
func (o *testOracle) GetTableSchema(context.Context, models.ConnectionProfile, string) ([]models.ColumnInfo, error) {
	return nil, nil
}
func (o *testOracle) ExecuteQuery(_ context.Context, p models.ConnectionProfile, _ string, rows int) (*models.QueryResult, error) {
	o.queryRows, o.queryProfile = rows, p.ID
	return &models.QueryResult{}, nil
}
func (o *testOracle) ExecuteReadOnlyQuery(ctx context.Context, p models.ConnectionProfile, query string, rows int) (*models.QueryResult, error) {
	o.readOnly = true
	return o.ExecuteQuery(ctx, p, query, rows)
}
func (o *testOracle) ExecutePLSQL(context.Context, models.ConnectionProfile, string) (*models.PLSQLResult, error) {
	return nil, errors.New("connection failed")
}

func inspectionTestServer(ora *testOracle) *MCPServer {
	s := NewMCPServer(nil, ora, audit.NewAuditManager(100))
	s.inspectionProfile = &models.ConnectionProfile{ID: "pinned", Username: "reader"}
	return s
}

func TestJSONRPCValidationAndIDs(t *testing.T) {
	s := NewMCPServer(nil, nil, nil)
	for _, test := range []struct {
		request string
		code    int
	}{
		{`{`, -32700}, {`[]`, -32600}, {`null`, -32600},
		{`{"jsonrpc":"1.0","id":1,"method":"ping"}`, -32600},
		{`{"jsonrpc":"2.0","id":true,"method":"ping"}`, -32600},
		{`{"jsonrpc":"2.0","id":null,"method":"ping"}`, -32600},
		{`{"jsonrpc":"2.0","id":1.5,"method":"ping"}`, -32600},
		{`{"jsonrpc":"2.0","id":1,"method":"ping","params":[]}`, -32602},
	} {
		response := s.ProcessJSONRPC(context.Background(), []byte(test.request), "test")
		var parsed struct {
			Error *RPCError       `json:"error"`
			ID    json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(response, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.Error == nil || parsed.Error.Code != test.code {
			t.Errorf("%s: got %s", test.request, response)
		}
		if test.code != -32602 && string(parsed.ID) != "null" {
			t.Errorf("invalid request must return null id: %s", response)
		}
	}
	for _, id := range []string{`0`, `9007199254740993`, `"request"`} {
		response := s.ProcessJSONRPC(context.Background(), []byte(`{"jsonrpc":"2.0","method":"ping","id":`+id+`}`), "test")
		var parsed struct {
			ID json.RawMessage `json:"id"`
		}
		_ = json.Unmarshal(response, &parsed)
		if string(parsed.ID) != id {
			t.Errorf("request ID changed: %s -> %s", id, response)
		}
	}
	if response := s.ProcessJSONRPC(context.Background(), []byte(`{"jsonrpc":"2.0","method":"tools/call","params":{"name":"oracle_switch_connection"}}`), "test"); len(response) > 0 {
		t.Fatalf("notification received a response: %s", response)
	}
}

func TestInspectionToolsAndArguments(t *testing.T) {
	ora := &testOracle{}
	s := inspectionTestServer(ora)
	tools := s.getToolsList()
	if len(tools) != 3 {
		t.Fatalf("expected three inspection tools, got %d", len(tools))
	}
	for _, tool := range tools {
		if tool.Name == "oracle_execute_plsql" || tool.Name == "oracle_switch_connection" || tool.Name == "oracle_list_connections" {
			t.Errorf("exposed %s", tool.Name)
		}
		if tool.Annotations == nil {
			t.Errorf("missing annotations on %s", tool.Name)
		}
		if tool.Name == "oracle_query" && tool.Annotations.ReadOnlyHint {
			t.Fatal("query must not claim unconditional read-only behavior")
		}
	}
	for _, args := range []map[string]interface{}{
		{"query": "SELECT 1 FROM dual", "max_rows": float64(0)},
		{"query": "SELECT 1 FROM dual", "max_rows": float64(-1)},
		{"query": "SELECT 1 FROM dual", "max_rows": 1.5},
		{"query": "SELECT 1 FROM dual", "max_rows": "100"},
		{"query": "SELECT 1 FROM dual", "max_rows": float64(501)},
		{"query": "SELECT 1 FROM dual", "unexpected": true},
		{"query": 42}, {},
	} {
		if _, isErr := s.callTool(context.Background(), "oracle_query", args, "test"); !isErr {
			t.Errorf("accepted %v", args)
		}
	}
	result, isErr := s.callTool(context.Background(), "oracle_query", map[string]interface{}{"query": "SELECT COUNT(*) FROM EMP"}, "test")
	if isErr {
		t.Fatal(result)
	}
	if !ora.readOnly || ora.queryProfile != "pinned" || ora.queryRows != 100 {
		t.Fatalf("inspection query did not use pinned read-only path: %+v", ora)
	}
	if _, isErr := s.callTool(context.Background(), "oracle_switch_connection", map[string]interface{}{"connection_id": "other"}, "test"); !isErr {
		t.Fatal("inspection allowed switching")
	}
	if got := plsqlErrorText(nil, errors.New("connection failed")); !strings.Contains(got, "connection failed") {
		t.Fatal(got)
	}
}

func TestPinnedConnectionPolicyLimitsAndPLSQLFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	keyring.MockInit()
	cfg, err := storage.NewConfigManager()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"first", "second"} {
		if err := cfg.SaveConnection(models.ConnectionProfile{ID: id, Name: id}); err != nil {
			t.Fatal(err)
		}
	}
	policy := models.SecurityPolicy{Mode: "full", AllowPLSQL: true, MaxRows: 10}
	if err := cfg.SaveSecurityPolicy(policy); err != nil {
		t.Fatal(err)
	}
	ora := &testOracle{}
	s := NewMCPServer(cfg, ora, audit.NewAuditManager(100))
	if err := s.ConfigureInspection("missing"); err == nil {
		t.Fatal("inspection silently fell back from unknown profile")
	}
	if err := s.ConfigureInspection(""); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetActiveConnection("second"); err != nil {
		t.Fatal(err)
	}
	if got := s.activeConnection().ID; got != "first" {
		t.Fatalf("inspection connection changed: %s", got)
	}
	if _, isErr := s.callTool(context.Background(), "oracle_query", map[string]interface{}{"query": "SELECT 1 FROM dual"}, "test"); isErr {
		t.Fatal("inspection SELECT failed")
	}
	if ora.queryRows != 10 || !ora.readOnly || ora.queryProfile != "first" {
		t.Fatalf("pinned query limits: %+v", ora)
	}
	if _, isErr := s.callTool(context.Background(), "oracle_query", map[string]interface{}{"query": "DELETE FROM EMP"}, "test"); !isErr {
		t.Fatal("inspection inherited full write permissions")
	}
	if err := s.ConfigureInspection("second"); err != nil {
		t.Fatal(err)
	}
	if s.activeConnection().ID != "second" {
		t.Fatal("explicit connection was ignored")
	}

	normal := NewMCPServer(cfg, ora, audit.NewAuditManager(100))
	text, isErr := normal.callTool(context.Background(), "oracle_execute_plsql", map[string]interface{}{"block": "BEGIN NULL; END;"}, "test")
	if !isErr || !strings.Contains(text, "connection failed") {
		t.Fatalf("PL/SQL nil result error: %s", text)
	}
	if _, isErr := normal.callTool(context.Background(), "oracle_query", map[string]interface{}{"query": "SELECT 1 FROM dual", "max_rows": float64(11)}, "test"); !isErr {
		t.Fatal("accepted rows above policy")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestStdioFramesAndWriteErrors(t *testing.T) {
	s := NewMCPServer(nil, nil, nil)
	input := "{\n" + `{"jsonrpc":"2.0","id":0,"method":"ping"}` + "\n" + `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n"
	var output bytes.Buffer
	if err := s.runStdio(context.Background(), io.NopCloser(strings.NewReader(input)), &output); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two JSON-only response frames: %s", output.String())
	}
	for _, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Fatalf("non-JSON stdout: %s", line)
		}
	}
	if err := s.runStdio(context.Background(), io.NopCloser(strings.NewReader(input)), failingWriter{}); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("write failure lost: %v", err)
	}
}

func TestStdioCancellationAndTimeout(t *testing.T) {
	for _, explicitCancel := range []bool{true, false} {
		t.Run(map[bool]string{true: "notification", false: "deadline"}[explicitCancel], func(t *testing.T) {
			ora := &testOracle{started: make(chan struct{}), finished: make(chan struct{})}
			s := inspectionTestServer(ora)
			timeout := 100 * time.Millisecond
			if explicitCancel {
				timeout = 2 * time.Second
			}
			_ = s.SetRequestTimeout(timeout)
			input, client := io.Pipe()
			responseReader, output := io.Pipe()
			defer responseReader.Close()
			defer output.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- s.runStdio(ctx, input, output) }()
			if _, err := io.WriteString(client, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"oracle_list_tables"}}`+"\n"); err != nil {
				t.Fatal(err)
			}
			select {
			case <-ora.started:
			case <-ctx.Done():
				t.Fatal("operation did not start")
			}
			if explicitCancel {
				if _, err := io.WriteString(client, `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":1}}`+"\n"); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case <-ora.finished:
			case <-ctx.Done():
				t.Fatal("operation was not cancelled")
			}
			// Read concurrently so an uncancelled response cannot block the writer.
			frames := make(chan string, 2)
			go func() {
				scanner := bufio.NewScanner(responseReader)
				for scanner.Scan() {
					frames <- scanner.Text()
				}
				close(frames)
			}()
			_, _ = io.WriteString(client, `{"jsonrpc":"2.0","id":2,"method":"ping"}`+"\n")
			_ = client.Close()
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("transport did not finish")
			}
			_ = output.Close()
			var received []string
			for frame := range frames {
				received = append(received, frame)
			}
			want := 1
			if !explicitCancel {
				want = 2
			}
			if len(received) != want {
				t.Fatalf("expected %d responses, got %v", want, received)
			}
			if explicitCancel && !strings.Contains(received[0], `"id":2`) {
				t.Fatalf("cancelled request received a response: %v", received)
			}
			if !explicitCancel && !strings.Contains(received[0], "deadline exceeded") {
				t.Fatalf("timeout was not surfaced: %v", received)
			}
		})
	}
}

func TestStdioQueueLimitAndParentCancellation(t *testing.T) {
	ora := &testOracle{started: make(chan struct{}), finished: make(chan struct{})}
	s := inspectionTestServer(ora)
	input, client := io.Pipe()
	defer client.Close()
	var output bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.runStdio(ctx, input, &output) }()
	_, _ = io.WriteString(client, `{"jsonrpc":"2.0","id":0,"method":"tools/call","params":{"name":"oracle_list_tables"}}`+"\n")
	select {
	case <-ora.started:
	case <-ctx.Done():
		t.Fatal("operation did not start")
	}
	for id := 1; id <= maxPendingRequests; id++ {
		request, _ := json.Marshal(map[string]interface{}{"jsonrpc": "2.0", "id": id, "method": "ping"})
		if _, err := client.Write(append(request, '\n')); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = io.WriteString(client, `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":0}}`+"\n")
	_ = client.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("queue did not drain")
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != maxPendingRequests || !strings.Contains(lines[0], "queue is full") {
		t.Fatalf("queue limit not enforced: %s", output.String())
	}

	reader, writer := io.Pipe()
	defer writer.Close()
	parent, stop := context.WithCancel(context.Background())
	go func() { done <- s.runStdio(parent, reader, io.Discard) }()
	stop()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not unblock STDIO")
	}
}
