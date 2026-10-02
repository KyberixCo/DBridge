package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"oramcp/backend/audit"
	"oramcp/backend/models"
	"oramcp/backend/security"
	"oramcp/backend/storage"
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 message.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 message.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
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
	Name        string           `json:"name"`
	Description string           `json:"description"`
	InputSchema interface{}      `json:"inputSchema"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// ToolAnnotations describe a tool's behavior to MCP clients. They do not enforce policy.
type ToolAnnotations struct {
	ReadOnlyHint    bool `json:"readOnlyHint"`
	DestructiveHint bool `json:"destructiveHint"`
	IdempotentHint  bool `json:"idempotentHint"`
	OpenWorldHint   bool `json:"openWorldHint"`
}

// MCPServer coordinates HTTP/SSE and Stdio MCP transports.
type MCPServer struct {
	configMgr *storage.ConfigManager
	oracleMgr oracleClient
	auditMgr  *audit.AuditManager

	mu          sync.RWMutex
	httpServer  *http.Server
	listener    net.Listener
	running     bool
	port        int
	sseSessions map[string]chan string

	totalRequests   int64
	blockedRequests int64

	// These options are fixed before starting a transport.
	requestTimeout    time.Duration
	inspectionProfile *models.ConnectionProfile
}

type oracleClient interface {
	GetSchemaObjects(context.Context, models.ConnectionProfile) (*models.SchemaInfo, error)
	GetTableSchema(context.Context, models.ConnectionProfile, string) ([]models.ColumnInfo, error)
	ExecuteQuery(context.Context, models.ConnectionProfile, string, int) (*models.QueryResult, error)
	ExecuteReadOnlyQuery(context.Context, models.ConnectionProfile, string, int) (*models.QueryResult, error)
	ExecutePLSQL(context.Context, models.ConnectionProfile, string) (*models.PLSQLResult, error)
}

// NewMCPServer instantiates an MCP server.
func NewMCPServer(cfg *storage.ConfigManager, ora oracleClient, aud *audit.AuditManager) *MCPServer {
	return &MCPServer{
		configMgr:      cfg,
		oracleMgr:      ora,
		auditMgr:       aud,
		sseSessions:    make(map[string]chan string),
		requestTimeout: 30 * time.Second,
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

// ProcessJSONRPC processes an incoming JSON-RPC payload and returns the response.
func (s *MCPServer) ProcessJSONRPC(ctx context.Context, data []byte, clientInfo string) []byte {
	req, rpcErr := decodeRequest(data)
	if rpcErr != nil {
		return s.makeResponse(nil, nil, rpcErr)
	}
	// Notifications never execute tools or receive responses.
	if len(req.ID) == 0 {
		return nil
	}
	if len(req.Params) > 0 && strings.TrimSpace(string(req.Params))[0] != '{' {
		return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: "Params must be an object"})
	}
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout())
	defer cancel()
	if err := ctx.Err(); err != nil {
		return s.makeResponse(req.ID, nil, &RPCError{Code: -32000, Message: err.Error()})
	}

	atomic.AddInt64(&s.totalRequests, 1)

	switch req.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: "Invalid initialize params"})
			}
		}
		protocolVersion := "2025-03-26"
		if params.ProtocolVersion == "2024-11-05" {
			protocolVersion = params.ProtocolVersion
		}
		res := map[string]interface{}{
			"protocolVersion": protocolVersion,
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
			"instructions": s.serverInstructions(),
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
		if err := s.validateToolArguments(params.Name, params.Arguments); err != nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: err.Error()})
		}
		result, isErr := s.callTool(ctx, params.Name, params.Arguments, clientInfo)
		if err := ctx.Err(); err != nil {
			result, isErr = err.Error(), true
		}
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
		resources, err := s.getResourcesList(ctx)
		if err != nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32000, Message: err.Error()})
		}
		return s.makeResponse(req.ID, map[string]interface{}{"resources": resources}, nil)

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil || params.URI == "" {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: "Invalid resource URI"})
		}
		activeProfile := s.activeConnection()
		if activeProfile == nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: "Resource is not available"})
		}
		resources, err := s.getResourcesList(ctx)
		if err != nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32000, Message: "Could not list resources: " + err.Error()})
		}
		tableName, found := findResourceTable(resources, params.URI)
		if !found {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32602, Message: "Resource is not available"})
		}
		columns, err := s.oracleMgr.GetTableSchema(ctx, *activeProfile, tableName)
		if err != nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32000, Message: "Could not read resource: " + err.Error()})
		}
		contents, err := tableResourceContents(params.URI, tableName, columns)
		if err != nil {
			return s.makeResponse(req.ID, nil, &RPCError{Code: -32603, Message: "Could not encode resource"})
		}
		return s.makeResponse(req.ID, map[string]interface{}{"contents": contents}, nil)

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
	tools := []ToolDefinition{
		{
			Name:        "oracle_list_tables",
			Description: "Lists all tables, views, and stored procedures accessible to the current Oracle user.",
			Annotations: &ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "oracle_describe_table",
			Description: "Inspects schema of a table, returning columns, data types, nullability, and primary keys.",
			Annotations: &ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
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
			Annotations: &ToolAnnotations{DestructiveHint: true, OpenWorldHint: true},
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
						"description": "Maximum rows to retrieve (default: up to 100; capped by server policy)",
						"minimum":     1,
						"maximum":     s.rowLimit(),
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "oracle_execute_plsql",
			Annotations: &ToolAnnotations{DestructiveHint: true, OpenWorldHint: true},
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
			Annotations: &ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "oracle_switch_connection",
			Annotations: &ToolAnnotations{IdempotentHint: true},
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
	filtered := make([]ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		if s.inspectionProfile != nil && (tool.Name == "oracle_execute_plsql" || tool.Name == "oracle_switch_connection" || tool.Name == "oracle_list_connections") {
			continue
		}
		schema := tool.InputSchema.(map[string]interface{})
		schema["additionalProperties"] = false
		if tool.Name == "oracle_query" && s.inspectionProfile != nil {
			tool.Description = "Executes one SELECT/CTE on the pinned inspection connection in an Oracle read-only transaction. Direct routine calls are restricted; database links and PL/SQL are disabled. Requires a least-privilege Oracle user."
			// Views can call autonomous routines; do not promise unconditional read-only behavior.
		}
		filtered = append(filtered, tool)
	}
	return filtered
}

func (s *MCPServer) callTool(ctx context.Context, name string, args map[string]interface{}, clientInfo string) (string, bool) {
	if err := s.validateToolArguments(name, args); err != nil {
		return err.Error(), true
	}
	activeProfile := s.activeConnection()
	policy := s.securityPolicy()

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
		if err := s.configMgr.SetActiveConnection(found.ID); err != nil {
			return fmt.Sprintf("Could not persist active connection: %v", err), true
		}
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

		maxRows := min(100, s.rowLimit())
		if mr, ok := args["max_rows"].(float64); ok {
			maxRows = int(min(mr, float64(s.rowLimit())))
		}

		// Security Validation
		allowed, reason := security.ValidateQuery(query, policy)
		if allowed && s.inspectionProfile != nil {
			allowed, reason = security.ValidateInspectionQuery(query)
		}
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
		var result *models.QueryResult
		var err error
		if s.inspectionProfile != nil || policy.Mode == "read_only" || policy.Mode == "" {
			result, err = s.oracleMgr.ExecuteReadOnlyQuery(ctx, *activeProfile, query, maxRows)
		} else {
			result, err = s.oracleMgr.ExecuteQuery(ctx, *activeProfile, query, maxRows)
		}
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
			return plsqlErrorText(result, err), true
		}

		s.auditMgr.Record(entry)
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data), false

	default:
		return fmt.Sprintf("Unknown tool '%s'", name), true
	}
}

func (s *MCPServer) getResourcesList(ctx context.Context) ([]map[string]interface{}, error) {
	activeProfile := s.activeConnection()
	if activeProfile == nil {
		return []map[string]interface{}{}, nil
	}

	info, err := s.oracleMgr.GetSchemaObjects(ctx, *activeProfile)
	if err != nil {
		return nil, err
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
	return resources, nil
}

// findResourceTable accepts only a URI currently advertised by resources/list.
func findResourceTable(resources []map[string]interface{}, uri string) (string, bool) {
	for _, resource := range resources {
		if resource["uri"] == uri {
			name, ok := resource["name"].(string)
			return name, ok
		}
	}
	return "", false
}

func tableResourceContents(uri, tableName string, columns []models.ColumnInfo) ([]map[string]interface{}, error) {
	metadata, err := json.Marshal(map[string]interface{}{"table": tableName, "columns": columns})
	if err != nil {
		return nil, err
	}
	return []map[string]interface{}{{
		"uri": uri, "mimeType": "application/json", "text": string(metadata),
	}}, nil
}
