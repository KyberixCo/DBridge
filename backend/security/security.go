package security

import (
	"fmt"
	"regexp"
	"strings"

	"oramcp/backend/models"
)

var (
	// Regex to remove single line comments: -- comment
	singleLineCommentRegex = regexp.MustCompile(`--[^\r\n]*`)
	// Regex to remove multi-line comments: /* comment */
	multiLineCommentRegex = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	// Regex to remove string literals: 'text' (handles escaped quotes '')
	stringLiteralRegex = regexp.MustCompile(`'([^']|'')*'`)
)

// SanitizeSQL removes comments and string literals so keyword detection only operates on SQL tokens.
func SanitizeSQL(sql string) string {
	// 1. Remove comments
	cleaned := multiLineCommentRegex.ReplaceAllString(sql, " ")
	cleaned = singleLineCommentRegex.ReplaceAllString(cleaned, " ")

	// 2. Remove string literals to avoid false positives (e.g. WHERE status = 'DELETED')
	cleaned = stringLiteralRegex.ReplaceAllString(cleaned, "''")

	return strings.TrimSpace(cleaned)
}

// ExtractTokens returns uppercase word tokens from sanitized SQL.
func ExtractTokens(cleanedSQL string) []string {
	wordRegex := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
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

	sanitized := SanitizeSQL(trimmed)
	tokens := ExtractTokens(sanitized)
	if len(tokens) == 0 {
		return false, "No valid SQL tokens found in query"
	}

	// 1. Read-Only Mode enforcement (Default): Only SELECT or WITH (CTE) queries allowed
	if policy.Mode == "read_only" || policy.Mode == "" {
		firstToken := tokens[0]
		if firstToken != "SELECT" && firstToken != "WITH" && firstToken != "EXPLAIN" {
			return false, fmt.Sprintf("Query blocked: Read-Only mode only permits SELECT statements (found '%s')", firstToken)
		}

		// Also check against any write keywords in the tokens
		readOnlyBlocked := []string{
			"INSERT", "UPDATE", "DELETE", "DROP", "ALTER",
			"TRUNCATE", "CREATE", "GRANT", "REVOKE", "MERGE",
			"RENAME", "EXEC", "EXECUTE", "INTO",
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
	sanitized := SanitizeSQL(trimmed)
	tokens := ExtractTokens(sanitized)

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
