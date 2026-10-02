package oracle

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"testing"

	"oramcp/backend/models"
)

type inspectionConnector struct{ conn *inspectionConn }

func (c inspectionConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (c inspectionConnector) Driver() driver.Driver                        { return inspectionDriver{} }

type inspectionDriver struct{}

func (inspectionDriver) Open(string) (driver.Conn, error) { return nil, errors.New("use connector") }

type inspectionConn struct {
	events         []string
	rejectReadOnly bool
}

func (c *inspectionConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c *inspectionConn) Close() error               { return nil }
func (c *inspectionConn) Ping(context.Context) error { return nil }
func (c *inspectionConn) Begin() (driver.Tx, error) {
	c.events = append(c.events, "begin")
	return inspectionTx{c}, nil
}
func (c *inspectionConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c.Begin()
}
func (c *inspectionConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.events = append(c.events, query)
	if c.rejectReadOnly {
		return nil, errors.New("read-only enforcement failed")
	}
	return driver.RowsAffected(0), nil
}
func (c *inspectionConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	c.events = append(c.events, "query")
	return &inspectionRows{conn: c}, nil
}

type inspectionTx struct{ conn *inspectionConn }

func (tx inspectionTx) Commit() error { tx.conn.events = append(tx.conn.events, "commit"); return nil }
func (tx inspectionTx) Rollback() error {
	tx.conn.events = append(tx.conn.events, "rollback")
	return nil
}

type inspectionRows struct {
	conn  *inspectionConn
	count int
}

func (r *inspectionRows) Columns() []string { return []string{"ID"} }
func (r *inspectionRows) Close() error {
	r.conn.events = append(r.conn.events, "rows close")
	return nil
}
func (r *inspectionRows) Next(values []driver.Value) error {
	if r.count >= 3 {
		return io.EOF
	}
	r.count++
	values[0] = int64(r.count)
	return nil
}

func TestInspectionTransactionEnforcement(t *testing.T) {
	for _, fail := range []bool{false, true} {
		conn := &inspectionConn{rejectReadOnly: fail}
		db := sql.OpenDB(inspectionConnector{conn})
		defer db.Close()
		cm := NewClientManager()
		cm.pools["reader"] = db
		result, err := cm.ExecuteReadOnlyQuery(context.Background(), models.ConnectionProfile{ID: "reader"}, "SELECT ID FROM EMP", 1)
		if fail {
			if err == nil || result != nil {
				t.Fatal("query ran without enforcing read-only transaction")
			}
			if want := []string{"begin", "SET TRANSACTION READ ONLY", "rollback"}; !reflect.DeepEqual(conn.events, want) {
				t.Fatalf("failure path: %v", conn.events)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if result.RowCount != 1 {
				t.Fatalf("row limit ignored: %+v", result)
			}
			if want := []string{"begin", "SET TRANSACTION READ ONLY", "query", "rows close", "rollback"}; !reflect.DeepEqual(conn.events, want) {
				t.Fatalf("transaction lifecycle: %v", conn.events)
			}
		}
	}
}

func TestPoolAcquisitionHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewClientManager().GetDBContext(ctx, models.ConnectionProfile{ID: "reader"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context was ignored: %v", err)
	}
}
