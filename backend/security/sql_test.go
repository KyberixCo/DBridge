package security

import (
	"testing"

	"oramcp/backend/models"
)

func TestSQLLexicalBoundaries(t *testing.T) {
	for _, query := range []string{
		"SELECT '--' AS message FROM dual; DELETE FROM users",
		"SELECT '/*' AS message FROM dual; DELETE FROM users -- */",
		"SELECT 1 FROM dual; SELECT 2 FROM dual",
		"SELECT 1 FROM dual;;",
		"SELECT 'unterminated FROM dual",
		"SELECT 1 FROM dual /* unterminated",
		"SELECT seq.NEXTVAL FROM dual",
		`SELECT seq."NEXTVAL" FROM dual`,
		"SELECT * FROM users FOR UPDATE",
		"EXPLAIN PLAN FOR SELECT 1 FROM dual",
		"WITH FUNCTION dangerous RETURN NUMBER IS BEGIN RETURN 1; END; SELECT dangerous FROM dual",
	} {
		if allowed, reason := ValidateQuery(query, models.DefaultSecurityPolicy()); allowed {
			t.Errorf("accepted %q (%s)", query, reason)
		}
	}
	for _, query := range []string{
		"SELECT '-- /* DELETE */' FROM dual; -- safe trailing comment",
		"SELECT q'[it's -- /* DELETE; */]' FROM dual",
		"SELECT q'{DROP; 'quoted'}' FROM dual;",
		"SELECT 'it''s safe' FROM dual",
		`SELECT "UPDATE" FROM "DELETE"`,
	} {
		if allowed, reason := ValidateQuery(query, models.DefaultSecurityPolicy()); !allowed {
			t.Errorf("rejected %q: %s", query, reason)
		}
	}
	full := models.SecurityPolicy{Mode: "full"}
	if allowed, _ := ValidateQuery("UPDATE users SET id = 1; DELETE FROM users", full); allowed {
		t.Fatal("full mode accepted a script")
	}
	if allowed, reason := ValidateQuery("UPDATE users SET id = 1;", full); !allowed {
		t.Fatalf("full mode single statement: %s", reason)
	}
}

func TestInspectionSQL(t *testing.T) {
	for _, query := range []string{
		"SELECT COUNT(*), NVL(MAX(salary), 0) FROM employees",
		"WITH summary AS (SELECT id FROM users) SELECT * FROM summary WHERE id IN (1,2)",
		"SELECT '-- @remote dangerous()' FROM dual",
		`SELECT "ID" FROM "HR"."EMP"`,
	} {
		if allowed, reason := ValidateInspectionQuery(query); !allowed {
			t.Errorf("rejected %q: %s", query, reason)
		}
	}
	for _, query := range []string{
		"SELECT dangerous() FROM dual",
		"SELECT UTL_HTTP.REQUEST('https://example.com') FROM dual",
		"SELECT pkg . COUNT(*) FROM dual",
		`SELECT "pkg".COUNT(*) FROM dual`,
		`SELECT "dangerous"() FROM dual`,
		"SELECT * FROM employees@remote",
	} {
		if allowed, _ := ValidateInspectionQuery(query); allowed {
			t.Errorf("accepted %q", query)
		}
	}
}
