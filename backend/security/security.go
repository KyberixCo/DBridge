package security

import (
	"fmt"
	"regexp"
	"strings"

	"oramcp/backend/models"
)

var (
	wordRegex           = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_$#]*`)
	functionCallRegex   = regexp.MustCompile(`((?:"_"|[a-zA-Z_][a-zA-Z0-9_$#]*)(?:\s*\.\s*(?:"_"|[a-zA-Z_][a-zA-Z0-9_$#]*))*)\s*\(`)
	inspectionFunctions = map[string]bool{
		// SQL syntax followed by parentheses, scalar functions and aggregates.
		"AS": true, "IN": true, "EXISTS": true, "OVER": true, "AND": true, "OR": true, "NOT": true,
		"SELECT": true, "FROM": true, "WHERE": true, "ON": true, "BY": true, "HAVING": true,
		"COUNT": true, "SUM": true, "MIN": true, "MAX": true, "AVG": true,
		"NVL": true, "NVL2": true, "COALESCE": true, "NULLIF": true, "DECODE": true,
		"UPPER": true, "LOWER": true, "TRIM": true, "LTRIM": true, "RTRIM": true,
		"LENGTH": true, "SUBSTR": true, "INSTR": true, "REPLACE": true, "CONCAT": true,
		"ABS": true, "ROUND": true, "TRUNC": true, "CEIL": true, "FLOOR": true, "MOD": true,
		"TO_CHAR": true, "TO_DATE": true, "TO_TIMESTAMP": true, "TO_NUMBER": true,
		"CAST": true, "EXTRACT": true, "NUMBER": true, "VARCHAR2": true, "CHAR": true,
		"ROW_NUMBER": true, "RANK": true, "DENSE_RANK": true, "LAG": true, "LEAD": true,
	}
)

// SanitizeSQL removes comments and string literals so keyword detection only operates on SQL tokens.
func SanitizeSQL(sql string) string {
	cleaned, _ := scanSQL(sql)
	return cleaned
}

// ExtractTokens returns uppercase word tokens from sanitized SQL.
func ExtractTokens(cleanedSQL string) []string {
	matches := wordRegex.FindAllString(cleanedSQL, -1)
	tokens := make([]string, len(matches))
	for i, m := range matches {
		tokens[i] = strings.ToUpper(m)
	}
	return tokens
}

// ValidateQuery evaluates a SQL query against the given SecurityPolicy.
func ValidateQuery(query string, policy models.SecurityPolicy) (bool, string) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return false, "Query cannot be empty"
	}

	sanitized, err := scanSQL(trimmed)
	if err != nil {
		return false, err.Error()
	}
	if !singleStatement(sanitized) {
		return false, "Only one SQL statement is permitted"
	}
	tokens := ExtractTokens(sanitized)
	if len(tokens) == 0 {
		return false, "No valid SQL tokens found in query"
	}

	// 1. Read-Only Mode enforcement (Default): Only SELECT or WITH (CTE) queries allowed
	if policy.Mode == "read_only" || policy.Mode == "" {
		firstToken := tokens[0]
		if firstToken != "SELECT" && firstToken != "WITH" {
			return false, fmt.Sprintf("Query blocked: Read-Only mode only permits SELECT statements (found '%s')", firstToken)
		}

		// Also check against any write keywords in the tokens
		readOnlyBlocked := []string{
			"INSERT", "UPDATE", "DELETE", "DROP", "ALTER",
			"TRUNCATE", "CREATE", "GRANT", "REVOKE", "MERGE",
			"RENAME", "EXEC", "EXECUTE", "INTO",
			"BEGIN", "DECLARE", "FUNCTION", "PROCEDURE", "PRAGMA", "CALL",
			"COMMIT", "ROLLBACK", "SAVEPOINT", "LOCK", "NEXTVAL",
		}

		for _, token := range tokens {
			for _, blocked := range readOnlyBlocked {
				if token == blocked {
					return false, fmt.Sprintf("Query blocked: Forbidden token '%s' detected in Read-Only mode", token)
				}
			}
		}

		return true, ""
	}

	// 2. Custom or Full Mode: Check against configured BlockedKeywords
	if len(policy.BlockedKeywords) > 0 {
		blockedMap := make(map[string]bool)
		for _, kw := range policy.BlockedKeywords {
			norm := strings.ToUpper(strings.TrimSpace(kw))
			if norm != "" {
				blockedMap[norm] = true
			}
		}

		for _, token := range tokens {
			if blockedMap[token] {
				return false, fmt.Sprintf("Query blocked: Prohibited keyword '%s' matches MCP security policy", token)
			}
		}
	}

	return true, ""
}

// ValidatePLSQL evaluates a PL/SQL block against the given SecurityPolicy.
func ValidatePLSQL(block string, policy models.SecurityPolicy) (bool, string) {
	if !policy.AllowPLSQL {
		return false, "PL/SQL execution via MCP is disabled by policy. Enable 'Allow PL/SQL' in OraMCP Security Settings."
	}

	trimmed := strings.TrimSpace(block)
	if trimmed == "" {
		return false, "PL/SQL block cannot be empty"
	}

	// For PL/SQL, inspect both sanitized tokens and full text to prevent dynamic SQL injection (e.g. EXECUTE IMMEDIATE 'DROP...')
	sanitized, err := scanSQL(trimmed)
	if err != nil {
		return false, err.Error()
	}
	tokens := ExtractTokens(sanitized)
	if len(tokens) == 0 || (tokens[0] != "BEGIN" && tokens[0] != "DECLARE") {
		return false, "PL/SQL must be an anonymous BEGIN or DECLARE block"
	}

	if len(policy.BlockedKeywords) > 0 {
		for _, kw := range policy.BlockedKeywords {
			norm := strings.ToUpper(strings.TrimSpace(kw))
			if norm == "" {
				continue
			}

			// Check tokens
			for _, token := range tokens {
				if token == norm {
					return false, fmt.Sprintf("PL/SQL blocked: Forbidden keyword '%s' found in block", token)
				}
			}

			// In PL/SQL, also check word boundary in raw block (to catch EXECUTE IMMEDIATE 'DROP ...')
			pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(norm) + `\b`)
			if pattern.MatchString(trimmed) {
				return false, fmt.Sprintf("PL/SQL blocked: Forbidden keyword '%s' detected (including dynamic SQL)", norm)
			}
		}
	}

	return true, ""
}
