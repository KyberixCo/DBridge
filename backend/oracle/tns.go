package oracle

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"oramcp/backend/models"
)

// DetectTNSFiles searches standard environment and filesystem locations for tnsnames.ora files.
func DetectTNSFiles() []string {
	var candidates []string

	// 1. TNS_ADMIN environment variable
	if tnsAdmin := os.Getenv("TNS_ADMIN"); tnsAdmin != "" {
		if fi, err := os.Stat(tnsAdmin); err == nil {
			if fi.IsDir() {
				candidates = append(candidates, filepath.Join(tnsAdmin, "tnsnames.ora"))
			} else {
				candidates = append(candidates, tnsAdmin)
			}
		}
	}

	// 2. ORACLE_HOME environment variable
	if oracleHome := os.Getenv("ORACLE_HOME"); oracleHome != "" {
		candidates = append(candidates, filepath.Join(oracleHome, "network", "admin", "tnsnames.ora"))
	}

	// 3. User Home Directory
	if userHome, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(userHome, "tnsnames.ora"),
			filepath.Join(userHome, ".oracle", "tnsnames.ora"),
			filepath.Join(userHome, "oracle", "network", "admin", "tnsnames.ora"),
		)
	}

	// 4. Platform specific default system paths
	candidates = append(candidates,
		"/etc/tnsnames.ora",
		"/etc/oracle/tnsnames.ora",
		"/opt/oracle/network/admin/tnsnames.ora",
		"/Library/Oracle/network/admin/tnsnames.ora",
		`C:\oracle\network\admin\tnsnames.ora`,
		`C:\app\client\network\admin\tnsnames.ora`,
	)

	// Filter only existing readable files and deduplicate
	seen := make(map[string]bool)
	var found []string

	for _, path := range candidates {
		cleaned := filepath.Clean(path)
		if seen[cleaned] {
			continue
		}
		if fi, err := os.Stat(cleaned); err == nil && !fi.IsDir() {
			seen[cleaned] = true
			found = append(found, cleaned)
		}
	}

	return found
}

// ParseTNSFile reads and parses a tnsnames.ora file from disk.
func ParseTNSFile(filePath string) ([]models.TNSEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tnsnames file %s: %w", filePath, err)
	}
	return ParseTNSContent(string(data))
}

// Regex patterns for parsing elements within a TNS descriptor
var (
	hostRegexp     = regexp.MustCompile(`(?i)\(\s*HOST\s*=\s*([^)\s]+)\s*\)`)
	portRegexp     = regexp.MustCompile(`(?i)\(\s*PORT\s*=\s*([0-9]+)\s*\)`)
	protocolRegexp = regexp.MustCompile(`(?i)\(\s*PROTOCOL\s*=\s*([^)\s]+)\s*\)`)
	serviceRegexp  = regexp.MustCompile(`(?i)\(\s*SERVICE_NAME\s*=\s*([^)\s]+)\s*\)`)
	sidRegexp      = regexp.MustCompile(`(?i)\(\s*SID\s*=\s*([^)\s]+)\s*\)`)
)

// ParseTNSContent parses raw string content of a tnsnames.ora file.
func ParseTNSContent(content string) ([]models.TNSEntry, error) {
	// Step 1: Strip comments line by line
	var cleanLines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		// Remove comments starting with #
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	joined := strings.Join(cleanLines, " ")
	if len(joined) == 0 {
		return nil, nil
	}

	// Step 2: Parse entries using parenthesis balancing
	var entries []models.TNSEntry
	n := len(joined)
	i := 0

	for i < n {
		// Look for '='
		eqIdx := strings.Index(joined[i:], "=")
		if eqIdx == -1 {
			break
		}
		eqIdx += i

		rawAliasPart := strings.TrimSpace(joined[i:eqIdx])
		if rawAliasPart == "" {
			i = eqIdx + 1
			continue
		}

		// Look for opening '(' after '='
		openIdx := strings.Index(joined[eqIdx+1:], "(")
		if openIdx == -1 {
			break
		}
		openIdx += eqIdx + 1

		// Verify nothing invalid between '=' and '('
		inBetween := strings.TrimSpace(joined[eqIdx+1 : openIdx])
		if inBetween != "" {
			i = openIdx
			continue
		}

		// Count matching parentheses
		depth := 1
		j := openIdx + 1
		for j < n && depth > 0 {
			if joined[j] == '(' {
				depth++
			} else if joined[j] == ')' {
				depth--
			}
			j++
		}

		if depth != 0 {
			// Unbalanced parentheses
			break
		}

		descriptorBody := joined[openIdx:j]

		// Parse the aliases (can be comma-separated, e.g. "ORCL, ORCL.WORLD")
		aliases := strings.Split(rawAliasPart, ",")
		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}

			entry := parseTNSDescriptor(alias, descriptorBody)
			entries = append(entries, entry)
		}

		i = j
	}

	return entries, nil
}

func parseTNSDescriptor(alias string, descriptor string) models.TNSEntry {
	entry := models.TNSEntry{
		Alias:    alias,
		Port:     1521,
		Protocol: "TCP",
		Raw:      descriptor,
	}

	if match := hostRegexp.FindStringSubmatch(descriptor); len(match) > 1 {
		entry.Host = match[1]
	}

	if match := portRegexp.FindStringSubmatch(descriptor); len(match) > 1 {
		if p, err := strconv.Atoi(match[1]); err == nil && p > 0 {
			entry.Port = p
		}
	}

	if match := protocolRegexp.FindStringSubmatch(descriptor); len(match) > 1 {
		entry.Protocol = strings.ToUpper(match[1])
		if entry.Protocol == "TCPS" {
			entry.SSL = true
		}
	}

	if entry.Port == 2484 {
		entry.SSL = true
	}

	if match := serviceRegexp.FindStringSubmatch(descriptor); len(match) > 1 {
		entry.ServiceName = match[1]
		entry.IsSID = false
	}

	if match := sidRegexp.FindStringSubmatch(descriptor); len(match) > 1 {
		entry.SID = match[1]
		if entry.ServiceName == "" {
			entry.IsSID = true
		}
	}

	return entry
}
