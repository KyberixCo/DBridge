package security

import (
	"fmt"
	"strings"

	"oramcp/backend/models"
)

// scanSQL masks literals, quoted identifiers and comments in one pass. Unlike
// separate regex replacements, comment markers inside literals cannot hide SQL.
// Punctuation is retained so statement separators and database links are visible.
func scanSQL(sql string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(sql); {
		switch {
		case strings.HasPrefix(sql[i:], "--"):
			end := strings.IndexAny(sql[i:], "\r\n")
			if end < 0 {
				i = len(sql)
			} else {
				i += end
			}
			out.WriteByte(' ')
		case strings.HasPrefix(sql[i:], "/*"):
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 {
				return "", fmt.Errorf("unterminated SQL comment")
			}
			i += end + 4
			out.WriteByte(' ')
		case (sql[i] == 'q' || sql[i] == 'Q') && i+2 < len(sql) && sql[i+1] == '\'':
			close := sql[i+2]
			switch close {
			case '[':
				close = ']'
			case '{':
				close = '}'
			case '(':
				close = ')'
			case '<':
				close = '>'
			}
			if close == '\'' || close == ' ' || close == '\n' || close == '\r' || close == '\t' {
				return "", fmt.Errorf("invalid alternative SQL quote")
			}
			end := strings.Index(sql[i+3:], string([]byte{close, '\''}))
			if end < 0 {
				return "", fmt.Errorf("unterminated alternative SQL literal")
			}
			i += end + 5
			out.WriteString(" '' ")
		case sql[i] == '\'' || sql[i] == '"':
			quote := sql[i]
			i++
			start := i
			closed := false
			for i < len(sql) {
				if sql[i] == quote {
					i++
					if i < len(sql) && sql[i] == quote {
						i++
						continue
					}
					closed = true
					break
				}
				i++
			}
			if !closed {
				return "", fmt.Errorf("unterminated SQL literal or identifier")
			}
			if quote == '"' {
				if strings.EqualFold(sql[start:i-1], "NEXTVAL") {
					out.WriteString(" NEXTVAL ")
				} else {
					out.WriteString(` "_" `)
				}
			} else {
				out.WriteString(" '' ")
			}
		default:
			if sql[i] == 0 {
				return "", fmt.Errorf("SQL contains a NUL byte")
			}
			out.WriteByte(sql[i])
			i++
		}
	}
	return strings.TrimSpace(out.String()), nil
}

// singleStatement permits one optional trailing semicolon, never a script.
func singleStatement(cleaned string) bool {
	cleaned = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(cleaned), ";"))
	return !strings.Contains(cleaned, ";")
}

// ValidateInspectionQuery adds restrictions on direct routine calls and remote
// database links. Views and synonyms still require least-privilege DB grants.
func ValidateInspectionQuery(query string) (bool, string) {
	policy := models.DefaultSecurityPolicy()
	if allowed, reason := ValidateQuery(query, policy); !allowed {
		return false, reason
	}
	cleaned, _ := scanSQL(query)
	if strings.Contains(cleaned, "@") {
		return false, "Database links are disabled in inspection mode"
	}
	for _, match := range functionCallRegex.FindAllStringSubmatch(cleaned, -1) {
		name := strings.ToUpper(match[1])
		if strings.ContainsAny(name, `."`) || !inspectionFunctions[name] {
			return false, fmt.Sprintf("Routine or expression '%s' is not permitted in inspection mode", match[1])
		}
	}
	return true, ""
}
