package security

import (
	"testing"

	"oramcp/backend/models"
)

func TestValidateQuery_ReadOnly(t *testing.T) {
	policy := models.DefaultSecurityPolicy() // mode = read_only

	// Allowed queries
	allowedQueries := []string{
		"SELECT * FROM EMPLOYEES",
		"select id, name from users where status = 'ACTIVE'",
		"SELECT * FROM logs WHERE message = 'INSERT failed for user'",
		"/* DELETE comment */ SELECT count(*) FROM orders",
		"-- UPDATE comment\nSELECT * FROM dual",
		"WITH summary AS (SELECT dept_id, count(*) as cnt FROM emp GROUP BY dept_id) SELECT * FROM summary",
	}

	for _, q := range allowedQueries {
		allowed, reason := ValidateQuery(q, policy)
		if !allowed {
			t.Errorf("Expected query to be allowed: %s (reason: %s)", q, reason)
		}
	}

	// Blocked queries
	blockedQueries := []string{
		"INSERT INTO EMPLOYEES (NAME) VALUES ('Alice')",
		"UPDATE EMPLOYEES SET SALARY = 5000 WHERE ID = 1",
		"DELETE FROM EMPLOYEES WHERE ID = 1",
		"DROP TABLE TEMP_TABLE",
		"TRUNCATE TABLE LOGS",
		"ALTER TABLE EMPLOYEES ADD (AGE NUMBER)",
		"SELECT * INTO NEW_TABLE FROM OLD_TABLE",
	}

	for _, q := range blockedQueries {
		allowed, _ := ValidateQuery(q, policy)
		if allowed {
			t.Errorf("Expected query to be blocked: %s", q)
		}
	}
}

func TestValidateQuery_CustomBlocklist(t *testing.T) {
	policy := models.SecurityPolicy{
		Mode:            "custom",
		BlockedKeywords: []string{"DELETE", "DROP"},
		MaxRows:         500,
	}

	// Allowed
	allowed, _ := ValidateQuery("UPDATE users SET status = 1", policy)
	if !allowed {
		t.Errorf("UPDATE should be allowed in this custom policy")
	}

	// Blocked
	allowed, reason := ValidateQuery("DELETE FROM users", policy)
	if allowed {
		t.Errorf("DELETE should be blocked by custom policy, got: %v", reason)
	}
}

func TestValidatePLSQL(t *testing.T) {
	// Disabled by default
	policy := models.DefaultSecurityPolicy()
	allowed, _ := ValidatePLSQL("BEGIN NULL; END;", policy)
	if allowed {
		t.Errorf("PLSQL should be disabled by default")
	}

	// Enabled
	policy.AllowPLSQL = true
	allowed, _ = ValidatePLSQL("BEGIN DBMS_OUTPUT.PUT_LINE('Hello'); END;", policy)
	if !allowed {
		t.Errorf("PLSQL should be allowed when AllowPLSQL is true")
	}

	// Enabled but contains blocked keyword
	policy.BlockedKeywords = []string{"DROP"}
	allowed, _ = ValidatePLSQL("BEGIN EXECUTE IMMEDIATE 'DROP TABLE test'; END;", policy)
	if allowed {
		t.Errorf("PLSQL containing DROP should be blocked")
	}
}
