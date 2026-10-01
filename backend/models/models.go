package models

import "time"

// ConnectionProfile represents an Oracle DB connection configuration.
type ConnectionProfile struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	ServiceName string    `json:"serviceName"`
	SID         string    `json:"sid"`
	IsSID       bool      `json:"isSid"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	SSL         bool      `json:"ssl"`
	WalletPath  string    `json:"walletPath"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ConnectionTestResult contains the outcome of testing a database connection.
type ConnectionTestResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	ServerVersion string `json:"serverVersion"`
	LatencyMs     int64  `json:"latencyMs"`
}

// QueryResult holds tabular data returned by a SQL query.
type QueryResult struct {
	Columns     []string                 `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	RowCount    int                      `json:"rowCount"`
	ExecutionMs int64                    `json:"executionMs"`
	Error       string                   `json:"error,omitempty"`
}

// PLSQLResult holds execution output and DBMS_OUTPUT messages from a PL/SQL block.
type PLSQLResult struct {
	Success      bool     `json:"success"`
	Output       []string `json:"output"`
	RowsAffected int64    `json:"rowsAffected"`
	ExecutionMs  int64    `json:"executionMs"`
	Error        string   `json:"error,omitempty"`
}

// SchemaInfo lists schema objects.
type SchemaInfo struct {
	Tables     []string `json:"tables"`
	Views      []string `json:"views"`
	Procedures []string `json:"procedures"`
}

// ColumnInfo holds details about a table or view column.
type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"dataType"`
	DataLength   int    `json:"dataLength"`
	Nullable     bool   `json:"nullable"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
}

// TNSEntry represents a parsed entry from a tnsnames.ora file.
type TNSEntry struct {
	Alias       string `json:"alias"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	ServiceName string `json:"serviceName"`
	SID         string `json:"sid"`
	IsSID       bool   `json:"isSid"`
	SSL         bool   `json:"ssl"`
	Raw         string `json:"raw"`
}

// SecurityPolicy defines the permissions and restrictions applied to MCP tool calls.
type SecurityPolicy struct {
	Mode            string   `json:"mode"`            // "read_only", "custom", "full"
	BlockedKeywords []string `json:"blockedKeywords"` // e.g. ["INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE"]
	AllowPLSQL      bool     `json:"allowPlsql"`      // whether AI / MCP can execute PL/SQL blocks
	MaxRows         int      `json:"maxRows"`         // max rows returned by oracle_query
}

// DefaultSecurityPolicy returns a secure-by-default policy (SELECT only, blocked writes, disabled PL/SQL).
func DefaultSecurityPolicy() SecurityPolicy {
	return SecurityPolicy{
		Mode: "read_only",
		BlockedKeywords: []string{
			"INSERT", "UPDATE", "DELETE", "DROP", "ALTER",
			"TRUNCATE", "CREATE", "GRANT", "REVOKE", "MERGE",
			"RENAME", "EXEC", "EXECUTE",
		},
		AllowPLSQL: false,
		MaxRows:    500,
	}
}

// AuditLogEntry records an incoming MCP tool invocation.
type AuditLogEntry struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	ClientInfo  string    `json:"clientInfo"`
	ToolName    string    `json:"toolName"`
	Query       string    `json:"query"`
	Allowed     bool      `json:"allowed"`
	Reason      string    `json:"reason"`
	ExecutionMs int64     `json:"executionMs"`
	Error       string    `json:"error,omitempty"`
}

// MCPServerStatus provides status info about the local MCP server.
type MCPServerStatus struct {
	Running              bool   `json:"running"`
	Port                 int    `json:"port"`
	SSEUrl               string `json:"sseUrl"`
	HttpUrl              string `json:"httpUrl"`
	ActiveConnectionID   string `json:"activeConnectionId"`
	ActiveConnectionName string `json:"activeConnectionName"`
	TotalRequests        int64  `json:"totalRequests"`
	BlockedRequests      int64  `json:"blockedRequests"`
}
