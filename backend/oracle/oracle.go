package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	go_ora "github.com/sijms/go-ora/v2"

	"oramcp/backend/models"
)

// ClientManager manages Oracle database connection pools and operations.
type ClientManager struct {
	mu    sync.RWMutex
	pools map[string]*sql.DB
}

// NewClientManager creates an instance of ClientManager.
func NewClientManager() *ClientManager {
	return &ClientManager{
		pools: make(map[string]*sql.DB),
	}
}

// BuildConnectionString constructs a go-ora URL from a ConnectionProfile.
func BuildConnectionString(p models.ConnectionProfile) string {
	port := p.Port
	if port <= 0 {
		port = 1521
	}

	options := make(map[string]string)
	if p.SSL {
		options["SSL"] = "true"
	}
	if p.WalletPath != "" {
		options["wallet"] = p.WalletPath
	}

	service := p.ServiceName
	if p.IsSID && p.SID != "" {
		service = ""
		options["sid"] = p.SID
	}

	return go_ora.BuildUrl(p.Host, port, service, p.Username, p.Password, options)
}

// GetDB returns or initializes a cached *sql.DB pool for the profile.
func (cm *ClientManager) GetDB(p models.ConnectionProfile) (*sql.DB, error) {
	return cm.GetDBContext(context.Background(), p)
}

// GetDBContext includes connection establishment in the caller's deadline.
func (cm *ClientManager) GetDBContext(ctx context.Context, p models.ConnectionProfile) (*sql.DB, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if db, exists := cm.pools[p.ID]; exists {
		// Ping to ensure still valid
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := db.PingContext(pingCtx); err == nil {
			return db, nil
		}
		// If ping failed, close and recreate
		_ = db.Close()
		delete(cm.pools, p.ID)
	}

	connStr := BuildConnectionString(p)
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open oracle connection: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping oracle database: %w", err)
	}

	cm.pools[p.ID] = db
	return db, nil
}

// InvalidatePool closes and removes a cached pool.
func (cm *ClientManager) InvalidatePool(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if db, exists := cm.pools[id]; exists {
		_ = db.Close()
		delete(cm.pools, id)
	}
}

// TestConnection attempts to connect and queries the Oracle banner.
func (cm *ClientManager) TestConnection(p models.ConnectionProfile) models.ConnectionTestResult {
	start := time.Now()
	connStr := BuildConnectionString(p)

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return models.ConnectionTestResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to initialize driver: %v", err),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return models.ConnectionTestResult{
			Success:   false,
			Message:   fmt.Sprintf("Connection failed: %v", err),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}

	// Fetch Oracle version banner
	var banner string
	row := db.QueryRowContext(ctx, "SELECT BANNER FROM V$VERSION WHERE ROWNUM = 1")
	_ = row.Scan(&banner)
	if banner == "" {
		banner = "Oracle Database (Connected)"
	}

	return models.ConnectionTestResult{
		Success:       true,
		Message:       "Connection successful",
		ServerVersion: banner,
		LatencyMs:     time.Since(start).Milliseconds(),
	}
}

// ExecuteQuery executes a SELECT query and returns rows formatted as maps.
func (cm *ClientManager) ExecuteQuery(ctx context.Context, p models.ConnectionProfile, query string, maxRows int) (*models.QueryResult, error) {
	return cm.executeQuery(ctx, p, query, maxRows, false)
}

// ExecuteReadOnlyQuery uses an Oracle read-only transaction on a dedicated connection.
// Privileged autonomous routines can bypass transaction restrictions: use a least-privilege user.
func (cm *ClientManager) ExecuteReadOnlyQuery(ctx context.Context, p models.ConnectionProfile, query string, maxRows int) (*models.QueryResult, error) {
	return cm.executeQuery(ctx, p, query, maxRows, true)
}

func (cm *ClientManager) executeQuery(ctx context.Context, p models.ConnectionProfile, query string, maxRows int, readOnly bool) (*models.QueryResult, error) {
	if maxRows <= 0 {
		maxRows = 500
	}

	// Sanitize programmatic SQL: strip trailing semicolons, slashes, and whitespace
	query = strings.TrimSpace(query)
	for strings.HasSuffix(query, ";") || strings.HasSuffix(query, "/") {
		query = strings.TrimRight(query, ";/")
		query = strings.TrimSpace(query)
	}

	db, err := cm.GetDBContext(ctx, p)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	var rows *sql.Rows
	if readOnly {
		// go-ora does not support sql.TxOptions.ReadOnly. Set the mode explicitly
		// as the first statement in a transaction, and always roll it back.
		tx, beginErr := db.BeginTx(ctx, nil)
		if beginErr != nil {
			return nil, beginErr
		}
		defer tx.Rollback()
		if _, err = tx.ExecContext(ctx, "SET TRANSACTION READ ONLY"); err != nil {
			return nil, fmt.Errorf("could not enforce Oracle read-only transaction: %w", err)
		}
		rows, err = tx.QueryContext(ctx, query)
	} else {
		rows, err = db.QueryContext(ctx, query)
	}
	if err != nil {
		return &models.QueryResult{
			ExecutionMs: time.Since(start).Milliseconds(),
			Error:       err.Error(),
		}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to read columns: %w", err)
	}

	resultRows := make([]map[string]interface{}, 0)
	count := 0

	colCount := len(columns)
	for rows.Next() {
		if count >= maxRows {
			break
		}

		values := make([]interface{}, colCount)
		valuePtrs := make([]interface{}, colCount)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]interface{}, colCount)
		for i, col := range columns {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				rowMap[col] = string(v)
			case time.Time:
				rowMap[col] = v.Format("2006-01-02 15:04:05")
			default:
				rowMap[col] = v
			}
		}

		resultRows = append(resultRows, rowMap)
		count++
	}

	if err := rows.Err(); err != nil {
		return &models.QueryResult{
			Columns:     columns,
			Rows:        resultRows,
			RowCount:    len(resultRows),
			ExecutionMs: time.Since(start).Milliseconds(),
			Error:       err.Error(),
		}, err
	}

	return &models.QueryResult{
		Columns:     columns,
		Rows:        resultRows,
		RowCount:    len(resultRows),
		ExecutionMs: time.Since(start).Milliseconds(),
	}, nil
}

// ExecutePLSQL runs a PL/SQL block and reads DBMS_OUTPUT buffer.
func (cm *ClientManager) ExecutePLSQL(ctx context.Context, p models.ConnectionProfile, block string) (*models.PLSQLResult, error) {
	db, err := cm.GetDBContext(ctx, p)
	if err != nil {
		return nil, err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain dedicated connection: %w", err)
	}
	defer conn.Close()

	start := time.Now()

	// 1. Enable DBMS_OUTPUT
	_, err = conn.ExecContext(ctx, "BEGIN DBMS_OUTPUT.ENABLE(1000000); END;")
	if err != nil {
		// Ignore error if DBMS_OUTPUT is not available in user schema
	}

	// 2. Format PL/SQL block if not wrapped
	trimmed := strings.TrimSpace(block)
	plsqlToExec := trimmed
	if !strings.HasPrefix(strings.ToUpper(trimmed), "BEGIN") &&
		!strings.HasPrefix(strings.ToUpper(trimmed), "DECLARE") {
		plsqlToExec = fmt.Sprintf("BEGIN\n%s\nEND;", trimmed)
	}

	// 3. Execute block
	res, execErr := conn.ExecContext(ctx, plsqlToExec)
	var rowsAffected int64
	if res != nil {
		rowsAffected, _ = res.RowsAffected()
	}

	// 4. Retrieve DBMS_OUTPUT lines
	outputLines := make([]string, 0)
	for i := 0; i < 2000; i++ {
		var line string
		var status int64
		_, err := conn.ExecContext(ctx, "BEGIN DBMS_OUTPUT.GET_LINE(:1, :2); END;",
			go_ora.Out{Dest: &line, Size: 32767},
			go_ora.Out{Dest: &status},
		)
		if err != nil || status != 0 {
			break
		}
		outputLines = append(outputLines, line)
	}

	execMs := time.Since(start).Milliseconds()

	if execErr != nil {
		return &models.PLSQLResult{
			Success:      false,
			Output:       outputLines,
			RowsAffected: rowsAffected,
			ExecutionMs:  execMs,
			Error:        execErr.Error(),
		}, execErr
	}

	return &models.PLSQLResult{
		Success:      true,
		Output:       outputLines,
		RowsAffected: rowsAffected,
		ExecutionMs:  execMs,
	}, nil
}

// GetSchemaObjects retrieves tables, views, and procedures belonging to the user.
func (cm *ClientManager) GetSchemaObjects(ctx context.Context, p models.ConnectionProfile) (*models.SchemaInfo, error) {
	db, err := cm.GetDBContext(ctx, p)
	if err != nil {
		return nil, err
	}

	info := &models.SchemaInfo{
		Tables:     make([]string, 0),
		Views:      make([]string, 0),
		Procedures: make([]string, 0),
	}

	info.Tables, err = queryStrings(ctx, db, "SELECT TABLE_NAME FROM USER_TABLES ORDER BY TABLE_NAME")
	if err != nil {
		return nil, err
	}
	info.Views, err = queryStrings(ctx, db, "SELECT VIEW_NAME FROM USER_VIEWS ORDER BY VIEW_NAME")
	if err != nil {
		return nil, err
	}
	info.Procedures, err = queryStrings(ctx, db, "SELECT OBJECT_NAME FROM USER_OBJECTS WHERE OBJECT_TYPE IN ('PROCEDURE', 'FUNCTION', 'PACKAGE') ORDER BY OBJECT_NAME")
	if err != nil {
		return nil, err
	}

	return info, nil
}

func queryStrings(ctx context.Context, db *sql.DB, query string, args ...interface{}) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// GetTableSchema returns column definitions and primary keys for a given table.
func (cm *ClientManager) GetTableSchema(ctx context.Context, p models.ConnectionProfile, tableName string) ([]models.ColumnInfo, error) {
	db, err := cm.GetDBContext(ctx, p)
	if err != nil {
		return nil, err
	}

	tableName = strings.ToUpper(strings.TrimSpace(tableName))

	// Get primary key columns
	pkQuery := `
		SELECT cols.COLUMN_NAME
		FROM USER_CONSTRAINTS cons
		JOIN USER_CONS_COLUMNS cols ON cons.CONSTRAINT_NAME = cols.CONSTRAINT_NAME
		WHERE cons.CONSTRAINT_TYPE = 'P' AND cons.TABLE_NAME = :1
	`
	pkMap := make(map[string]bool)
	primaryKeys, err := queryStrings(ctx, db, pkQuery, tableName)
	if err != nil {
		return nil, err
	}
	for _, col := range primaryKeys {
		pkMap[col] = true
	}

	// Get columns
	colQuery := `
		SELECT COLUMN_NAME, DATA_TYPE, NVL(DATA_LENGTH, 0), NULLABLE
		FROM USER_TAB_COLS
		WHERE TABLE_NAME = :1
		ORDER BY COLUMN_ID
	`
	rows, err := db.QueryContext(ctx, colQuery, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		var nullableStr string
		if err := rows.Scan(&col.Name, &col.DataType, &col.DataLength, &nullableStr); err != nil {
			return nil, err
		}
		col.Nullable = (nullableStr == "Y")
		col.IsPrimaryKey = pkMap[col.Name]
		columns = append(columns, col)
	}

	return columns, rows.Err()
}
