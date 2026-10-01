package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"oramcp/backend/models"
)

// DetectDBeaverFiles searches standard OS locations for DBeaver data-sources.json.
func DetectDBeaverFiles() []string {
	var candidates []string

	userHome, err := os.UserHomeDir()
	if err == nil {
		// macOS
		candidates = append(candidates,
			filepath.Join(userHome, "Library", "DBeaverData", "workspace6", "General", ".dbeaver", "data-sources.json"),
			filepath.Join(userHome, "Library", "Application Support", "DBeaverData", "workspace6", "General", ".dbeaver", "data-sources.json"),
		)

		// Linux & standard XDG
		candidates = append(candidates,
			filepath.Join(userHome, ".local", "share", "DBeaverData", "workspace6", "General", ".dbeaver", "data-sources.json"),
			filepath.Join(userHome, ".var", "app", "io.dbeaver.DBeaverCommunity", "data", "DBeaverData", "workspace6", "General", ".dbeaver", "data-sources.json"),
		)

		// Windows AppData
		if appData := os.Getenv("APPDATA"); appData != "" {
			candidates = append(candidates,
				filepath.Join(appData, "DBeaverData", "workspace6", "General", ".dbeaver", "data-sources.json"),
			)
		}
	}

	var found []string
	seen := make(map[string]bool)

	for _, p := range candidates {
		cleaned := filepath.Clean(p)
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

type dbeaverRoot struct {
	Connections map[string]dbeaverConnection `json:"connections"`
}

type dbeaverConnection struct {
	Name          string                    `json:"name"`
	Provider      string                    `json:"provider"`
	Driver        string                    `json:"driver"`
	Configuration dbeaverConnectionConfig   `json:"configuration"`
}

type dbeaverConnectionConfig struct {
	Host               string            `json:"host"`
	Port               string            `json:"port"`
	Database           string            `json:"database"`
	Server             string            `json:"server"`
	Url                string            `json:"url"`
	User               string            `json:"user"`
	ProviderProperties map[string]string `json:"provider-properties"`
}

// Regexes for JDBC URLs
var (
	jdbcThinSlash = regexp.MustCompile(`(?i)jdbc:oracle:thin:@\/\/(.+?):([0-9]+)\/(.+)`)
	jdbcThinColon = regexp.MustCompile(`(?i)jdbc:oracle:thin:@(.+?):([0-9]+):(.+)`)
)

// ParseDBeaverFile reads a data-sources.json and extracts all Oracle connections.
func ParseDBeaverFile(filePath string) ([]models.ConnectionProfile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read DBeaver file %s: %w", filePath, err)
	}

	return ParseDBeaverContent(data)
}

// ParseDBeaverContent parses raw JSON data from a DBeaver data-sources.json.
func ParseDBeaverContent(data []byte) ([]models.ConnectionProfile, error) {
	var root dbeaverRoot
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse DBeaver JSON: %w", err)
	}

	var profiles []models.ConnectionProfile

	for id, conn := range root.Connections {
		isOracle := strings.EqualFold(conn.Provider, "oracle") ||
			strings.Contains(strings.ToLower(conn.Driver), "oracle") ||
			strings.HasPrefix(strings.ToLower(conn.Configuration.Url), "jdbc:oracle:")

		if !isOracle {
			continue
		}

		cfg := conn.Configuration
		name := conn.Name
		if name == "" {
			name = "Oracle " + id
		}

		host := cfg.Host
		port := 1521
		if p, err := strconv.Atoi(cfg.Port); err == nil && p > 0 {
			port = p
		}

		username := cfg.User
		database := strings.TrimSpace(cfg.Database)
		isSid := false
		serviceName := ""
		sid := ""

		// Check DBeaver provider property: @dbeaver-sid-service@
		mode := strings.ToUpper(cfg.ProviderProperties["@dbeaver-sid-service@"])
		if mode == "SID" {
			isSid = true
			sid = database
		} else if mode == "SERVICE" {
			isSid = false
			serviceName = database
		} else {
			// Check URL if available
			if cfg.Url != "" {
				if match := jdbcThinSlash.FindStringSubmatch(cfg.Url); len(match) > 3 {
					host = match[1]
					if p, err := strconv.Atoi(match[2]); err == nil && p > 0 {
						port = p
					}
					serviceName = match[3]
					isSid = false
				} else if match := jdbcThinColon.FindStringSubmatch(cfg.Url); len(match) > 3 {
					host = match[1]
					if p, err := strconv.Atoi(match[2]); err == nil && p > 0 {
						port = p
					}
					sid = match[3]
					isSid = true
				}
			}

			// If still unresolved, default to ServiceName unless database explicitly matches common SID or user flagged
			if serviceName == "" && sid == "" {
				if strings.Contains(strings.ToUpper(database), "PDB") {
					serviceName = database
					isSid = false
				} else {
					serviceName = database
				}
			}
		}

		ssl := port == 2484 || strings.Contains(strings.ToLower(cfg.Url), "tcps")

		profiles = append(profiles, models.ConnectionProfile{
			ID:          id,
			Name:        name,
			Host:        host,
			Port:        port,
			ServiceName: serviceName,
			SID:         sid,
			IsSID:       isSid,
			Username:    username,
			SSL:         ssl,
		})
	}

	return profiles, nil
}
