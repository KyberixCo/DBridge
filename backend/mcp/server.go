package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"oramcp/backend/audit"
	"oramcp/backend/models"
	"oramcp/backend/oracle"
	"oramcp/backend/security"
	"oramcp/backend/storage"
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 message.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 message.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error payload.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolCallParams defines parameters received in tools/call.
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolDefinition defines an MCP tool schema.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// MCPServer coordinates HTTP/SSE and Stdio MCP transports.
type MCPServer struct {
	configMgr *storage.ConfigManager
	oracleMgr *oracle.ClientManager
	auditMgr  *audit.AuditManager

	mu          sync.RWMutex
	httpServer  *http.Server
	listener    net.Listener
	running     bool
	port        int
	sseSessions map[string]chan string

	totalRequests   int64
	blockedRequests int64
}

// NewMCPServer instantiates an MCP server.
func NewMCPServer(cfg *storage.ConfigManager, ora *oracle.ClientManager, aud *audit.AuditManager) *MCPServer {
	return &MCPServer{
		configMgr:   cfg,
		oracleMgr:   ora,
		auditMgr:    aud,
		sseSessions: make(map[string]chan string),
	}
}

// StartHTTP starts the SSE and HTTP JSON-RPC listener.
func (s *MCPServer) StartHTTP(port int) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}

	if port <= 0 {
		port = s.configMgr.GetMCPPort()
	}
	s.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", s.handleSSE)
	mux.HandleFunc("/message", s.handleMessage)
	mux.HandleFunc("/mcp", s.handleDirectMCP)
	mux.HandleFunc("/health", s.handleHealth)

	// Add CORS middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to bind MCP server port %d: %w", port, err)
	}

	s.listener = listener
	s.httpServer = &http.Server{Handler: handler}
	s.running = true
	s.mu.Unlock()

	go func() {
		_ = s.httpServer.Serve(listener)
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	return nil
}

// StopHTTP shuts down the HTTP server.
func (s *MCPServer) StopHTTP() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	s.running = false
	return err
}

// GetStatus returns the current runtime status of the MCP server.
func (s *MCPServer) GetStatus() models.MCPServerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeConn := s.configMgr.GetActiveConnection()
	activeName := "None"
	activeID := ""
	if activeConn != nil {
		activeName = activeConn.Name
		activeID = activeConn.ID
	}

	total := atomic.LoadInt64(&s.totalRequests)
	blocked := atomic.LoadInt64(&s.blockedRequests)

	return models.MCPServerStatus{
		Running:              s.running,
		Port:                 s.port,
		SSEUrl:               fmt.Sprintf("http://localhost:%d/sse", s.port),
		HttpUrl:              fmt.Sprintf("http://localhost:%d/mcp", s.port),
		ActiveConnectionID:   activeID,
		ActiveConnectionName: activeName,
		TotalRequests:        total,
		BlockedRequests:      blocked,
	}
}

// handleHealth returns server health info.
func (s *MCPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := s.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// handleSSE serves Server-Sent Events stream for MCP clients.
func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
	msgChan := make(chan string, 100)

	s.mu.Lock()
	s.sseSessions[sessionID] = msgChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.sseSessions, sessionID)
		close(msgChan)
		s.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Emit endpoint event
	fmt.Fprintf(w, "event: endpoint\ndata: /message?sessionId=%s\n\n", sessionID)
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Ping keep-alive
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// handleMessage receives client JSON-RPC commands and dispatches response via SSE.
func (s *MCPServer) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	s.mu.RLock()
	msgChan, exists := s.sseSessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	respBytes := s.ProcessJSONRPC(r.Context(), body, "sse:"+r.RemoteAddr)
	if len(respBytes) > 0 {
		select {
		case msgChan <- string(respBytes):
		default:
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// handleDirectMCP handles stateless direct JSON-RPC requests via POST.
func (s *MCPServer) handleDirectMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	respBytes := s.ProcessJSONRPC(r.Context(), body, "http:"+r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBytes)
}

// RunStdio reads from stdin and writes to stdout for CLI MCP mode.
func (s *MCPServer) RunStdio(ctx context.Context) error {
	fmt.Fprintln(os.Stderr, "dbridge: MCP stdio server listening...")
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		trimmed := strings.TrimSpace(string(line))
		if len(trimmed) == 0 {
			continue
		}
		resp := s.ProcessJSONRPC(ctx, []byte(trimmed), "stdio")
		if len(resp) > 0 {
			fmt.Printf("%s\n", string(resp))
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "dbridge: stdio error: %v\n", err)
		return err
	}
	return nil
}

// ProcessJSONRPC processes an incoming JSON-RPC payload and returns the response.
func (s *MCPServer) ProcessJSONRPC(ctx context.Context, data []byte, clientInfo string) []byte {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &RPCError{
				Code:    -32700,
				Message: "Parse error: " + err.Error(),
			},
		}
		out, _ := json.Marshal(resp)
		return out
	}

	atomic.AddInt64(&s.totalRequests, 1)

	switch req.Method {
	case "initialize":
		res := map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
				"resources": map[string]interface{}{
					"listChanged": false,
				},
				"prompts": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "dbridge-server",
				"version": "1.0.0",
			},
		}
		return s.makeResponse(req.ID, res, nil)

	case "notifications/initialized", "initialized":
		return nil

	case "ping":
		return s.makeResponse(req.ID, map[string]interface{}{}, nil)

	case "tools/list":
		tools := s.getToolsList()
		return s.makeResponse(req.ID, map[string]interface{}{"tools": tools}, nil)

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: "Invalid params"})
		}
		result, isErr := s.callTool(ctx, params.Name, params.Arguments, clientInfo)
		respData := map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
			"isError": isErr,
		}
		return s.makeResponse(req.ID, respData, nil)

	case "resources/list":
		resources := s.getResourcesList(ctx)
		return s.makeResponse(req.ID, map[string]interface{}{"resources": resources}, nil)

	case "resources/templates/list":
		return s.makeResponse(req.ID, map[string]interface{}{"resourceTemplates": []interface{}{}}, nil)

	case "prompts/list":
		return s.makeResponse(req.ID, map[string]interface{}{"prompts": []interface{}{}}, nil)

	case "logging/setLevel":
		return s.makeResponse(req.ID, map[string]interface{}{}, nil)

	default:
		// If it's a notification without ID, never return an error per JSON-RPC 2.0 specification
		if req.ID == nil {
			return nil
		}
		return s.makeResponse(req.ID, nil, &RPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method '%s' not found", req.Method),
		})
	}
}

func (s *MCPServer) makeResponse(id interface{}, result interface{}, err *RPCError) []byte {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
		Error:   err,
	}
	out, _ := json.Marshal(resp)
	return out
}

func (s *MCPServer) getToolsList() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "oracle_list_tables",
			Description: "Lists all tables, views, and stored procedures accessible to the current Oracle user.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "oracle_describe_table",
			Description: "Inspects schema of a table, returning columns, data types, nullability, and primary keys.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"table_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the table to describe (case-insensitive)",
					},
				},
				"required": []string{"table_name"},
			},
		},
		{
			Name:        "oracle_query",
			Description: "Executes a SQL query against Oracle DB. Subject to OraMCP security policy (defaults to SELECT only; blocks modifications like INSERT, UPDATE, DELETE).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL query to execute (e.g. 'SELECT * FROM EMPLOYEES FETCH FIRST 50 ROWS ONLY')",
					},
					"max_rows": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of rows to retrieve (default: 100, max: 1000)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "oracle_execute_plsql",
			Description: "Executes an anonymous PL/SQL block (e.g. 'BEGIN ... END;') and captures DBMS_OUTPUT. Requires 'Allow PL/SQL' to be enabled in OraMCP security settings.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"block": map[string]interface{}{
						"type":        "string",
						"description": "PL/SQL code block to execute (must start with BEGIN or DECLARE)",
					},
				},
				"required": []string{"block"},
			},
		},
		{
			Name:        "oracle_list_connections",
			Description: "Returns all saved Oracle database connection profiles and indicates which is currently active.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "oracle_switch_connection",
			Description: "Switches the active Oracle connection used by MCP tools.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"connection_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID or exact name of the connection profile to activate",
					},
				},
				"required": []string{"connection_id"},
			},
		},
	}
}

func (s *MCPServer) callTool(ctx context.Context, name string, args map[string]interface{}, clientInfo string) (string, bool) {
	activeProfile := s.configMgr.GetActiveConnection()
	policy := s.configMgr.GetSecurityPolicy()

	switch name {
	case "oracle_list_connections":
		conns := s.configMgr.GetConnections()
		type ConnSummary struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Host     string `json:"host"`
			User     string `json:"username"`
			IsActive bool   `json:"isActive"`
		}
		var summaries []ConnSummary
		for _, c := range conns {
			summaries = append(summaries, ConnSummary{
				ID:       c.ID,
				Name:     c.Name,
				Host:     c.Host,
				User:     c.Username,
				IsActive: activeProfile != nil && activeProfile.ID == c.ID,
			})
		}
		data, _ := json.MarshalIndent(summaries, "", "  ")
		return string(data), false

	case "oracle_switch_connection":
		target, _ := args["connection_id"].(string)
		if target == "" {
			return "Missing 'connection_id' argument", true
		}
		conns := s.configMgr.GetConnections()
		var found *models.ConnectionProfile
		for _, c := range conns {
			if c.ID == target || strings.EqualFold(c.Name, target) {
				found = &c
				break
			}
		}
		if found == nil {
			return fmt.Sprintf("Connection profile '%s' not found", target), true
		}
		_ = s.configMgr.SetActiveConnection(found.ID)
		return fmt.Sprintf("Switched active Oracle connection to '%s' (%s@%s)", found.Name, found.Username, found.Host), false

	case "oracle_list_tables":
		if activeProfile == nil {
			return "No active Oracle connection configured in OraMCP.", true
		}
		info, err := s.oracleMgr.GetSchemaObjects(ctx, *activeProfile)
		if err != nil {
			return fmt.Sprintf("Error fetching schema objects: %v", err), true
		}
		data, _ := json.MarshalIndent(info, "", "  ")
		return string(data), false

	case "oracle_describe_table":
		if activeProfile == nil {
			return "No active Oracle connection configured in OraMCP.", true
		}
		tblName, _ := args["table_name"].(string)
		if tblName == "" {
			return "Missing 'table_name' argument", true
		}
		cols, err := s.oracleMgr.GetTableSchema(ctx, *activeProfile, tblName)
		if err != nil {
			return fmt.Sprintf("Error describing table '%s': %v", tblName, err), true
		}
		if len(cols) == 0 {
			return fmt.Sprintf("Table or view '%s' not found or has no visible columns", tblName), true
		}
		data, _ := json.MarshalIndent(cols, "", "  ")
		return string(data), false

	case "oracle_query":
		query, _ := args["query"].(string)
		if query == "" {
			return "Missing 'query' argument", true
		}

		maxRows := 100
		if mr, ok := args["max_rows"]; ok {
			switch v := mr.(type) {
			case float64:
				maxRows = int(v)
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					maxRows = parsed
				}
			}
		}
		if policy.MaxRows > 0 && maxRows > policy.MaxRows {
			maxRows = policy.MaxRows
		}

		// Security Validation
		allowed, reason := security.ValidateQuery(query, policy)
		entry := models.AuditLogEntry{
			ID:         fmt.Sprintf("log-%d", time.Now().UnixNano()),
			Timestamp:  time.Now(),
			ClientInfo: clientInfo,
			ToolName:   "oracle_query",
			Query:      query,
			Allowed:    allowed,
			Reason:     reason,
		}

		if !allowed {
			atomic.AddInt64(&s.blockedRequests, 1)
			s.auditMgr.Record(entry)
			return fmt.Sprintf("SECURITY BLOCKED: %s", reason), true
		}

		if activeProfile == nil {
			entry.Error = "No active connection"
			s.auditMgr.Record(entry)
			return "No active Oracle connection configured in OraMCP.", true
		}

		start := time.Now()
		result, err := s.oracleMgr.ExecuteQuery(ctx, *activeProfile, query, maxRows)
		entry.ExecutionMs = time.Since(start).Milliseconds()

		if err != nil {
			entry.Error = err.Error()
			s.auditMgr.Record(entry)
			return fmt.Sprintf("SQL Execution Error: %v", err), true
		}

		s.auditMgr.Record(entry)
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data), false

	case "oracle_execute_plsql":
		block, _ := args["block"].(string)
		if block == "" {
			return "Missing 'block' argument", true
		}

		// Security Validation
		allowed, reason := security.ValidatePLSQL(block, policy)
		entry := models.AuditLogEntry{
			ID:         fmt.Sprintf("log-%d", time.Now().UnixNano()),
			Timestamp:  time.Now(),
			ClientInfo: clientInfo,
			ToolName:   "oracle_execute_plsql",
			Query:      block,
			Allowed:    allowed,
			Reason:     reason,
		}

		if !allowed {
			atomic.AddInt64(&s.blockedRequests, 1)
			s.auditMgr.Record(entry)
			return fmt.Sprintf("SECURITY BLOCKED: %s", reason), true
		}

		if activeProfile == nil {
			entry.Error = "No active connection"
			s.auditMgr.Record(entry)
			return "No active Oracle connection configured in OraMCP.", true
		}

		start := time.Now()
		result, err := s.oracleMgr.ExecutePLSQL(ctx, *activeProfile, block)
		entry.ExecutionMs = time.Since(start).Milliseconds()

		if err != nil {
			entry.Error = err.Error()
			s.auditMgr.Record(entry)
			return fmt.Sprintf("PL/SQL Execution Error: %v\nOutput:\n%s", err, strings.Join(result.Output, "\n")), true
		}

		s.auditMgr.Record(entry)
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data), false

	default:
		return fmt.Sprintf("Unknown tool '%s'", name), true
	}
}

func (s *MCPServer) getResourcesList(ctx context.Context) []map[string]interface{} {
	activeProfile := s.configMgr.GetActiveConnection()
	if activeProfile == nil {
		return []map[string]interface{}{}
	}

	info, err := s.oracleMgr.GetSchemaObjects(ctx, *activeProfile)
	if err != nil {
		return []map[string]interface{}{}
	}

	resources := make([]map[string]interface{}, 0, len(info.Tables))
	for _, tbl := range info.Tables {
		resources = append(resources, map[string]interface{}{
			"uri":         fmt.Sprintf("oracle://%s/table/%s", activeProfile.Username, tbl),
			"name":        tbl,
			"description": fmt.Sprintf("Oracle table %s", tbl),
			"mimeType":    "application/json",
		})
	}
	return resources
}
