package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"oramcp/backend/models"
)

// AppConfig is the root configuration structure saved on disk.
type AppConfig struct {
	ActiveConnectionID string                     `json:"activeConnectionId"`
	Connections        []models.ConnectionProfile `json:"connections"`
	SecurityPolicy     models.SecurityPolicy      `json:"securityPolicy"`
	MCPPort            int                        `json:"mcpPort"`
}

// ConfigManager handles loading and persisting configurations safely.
type ConfigManager struct {
	mu        sync.RWMutex
	filePath  string
	config    AppConfig
	secretMgr *SecretManager
}

// NewConfigManager initializes the configuration manager and loads existing configuration.
func NewConfigManager() (*ConfigManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".oramcp")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	cfgPath := filepath.Join(dir, "config.json")

	cm := &ConfigManager{
		filePath:  cfgPath,
		secretMgr: NewSecretManager(),
		config: AppConfig{
			Connections:    make([]models.ConnectionProfile, 0),
			SecurityPolicy: models.DefaultSecurityPolicy(),
			MCPPort:        8085,
		},
	}

	if err := cm.load(); err != nil {
		// If file doesn't exist, save defaults
		_ = cm.save()
	}

	return cm, nil
}

func (cm *ConfigManager) load() error {
	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return err
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if cfg.MCPPort == 0 {
		cfg.MCPPort = 8085
	}
	if cfg.SecurityPolicy.Mode == "" {
		cfg.SecurityPolicy = models.DefaultSecurityPolicy()
	}

	// Automatic Migration: check for any plain text passwords in config.json
	// and transfer them into the native OS Keychain, then scrub them from disk.
	migrated := false
	for i, c := range cfg.Connections {
		if c.Password != "" {
			_ = cm.secretMgr.SetPassword(c.ID, c.Password)
			cfg.Connections[i].Password = "" // scrub from memory/json
			migrated = true
		}
	}

	cm.config = cfg

	if migrated {
		_ = cm.save()
	}

	return nil
}

func (cm *ConfigManager) save() error {
	// Guarantee that passwords are never written to disk in plain text
	scrubbedConfig := cm.config
	scrubbedConns := make([]models.ConnectionProfile, len(cm.config.Connections))
	for i, c := range cm.config.Connections {
		scrubbed := c
		scrubbed.Password = "" // strip
		scrubbedConns[i] = scrubbed
	}
	scrubbedConfig.Connections = scrubbedConns

	data, err := json.MarshalIndent(scrubbedConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cm.filePath, data, 0600)
}

// GetConfig returns a copy of the configuration.
func (cm *ConfigManager) GetConfig() AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// GetConnections returns all saved connection profiles (with passwords scrubbed).
func (cm *ConfigManager) GetConnections() []models.ConnectionProfile {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	res := make([]models.ConnectionProfile, len(cm.config.Connections))
	for i, c := range cm.config.Connections {
		res[i] = c
		res[i].Password = "" // secure
	}
	return res
}

// SaveConnection creates or updates a connection profile and securely vaults password in OS Keychain.
func (cm *ConfigManager) SaveConnection(profile models.ConnectionProfile) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// If password was provided, store in OS Keychain
	if profile.Password != "" {
		if err := cm.secretMgr.SetPassword(profile.ID, profile.Password); err != nil {
			return err
		}
	}

	// Scrub password for local state
	sanitized := profile
	sanitized.Password = ""

	found := false
	for i, c := range cm.config.Connections {
		if c.ID == profile.ID {
			cm.config.Connections[i] = sanitized
			found = true
			break
		}
	}
	if !found {
		cm.config.Connections = append(cm.config.Connections, sanitized)
	}

	// If no active connection is set, set this one
	if cm.config.ActiveConnectionID == "" {
		cm.config.ActiveConnectionID = profile.ID
	}

	return cm.save()
}

// DeleteConnection removes a connection from config and purges its secret from OS Keychain.
func (cm *ConfigManager) DeleteConnection(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove secret from OS Keychain
	_ = cm.secretMgr.DeletePassword(id)

	filtered := make([]models.ConnectionProfile, 0, len(cm.config.Connections))
	for _, c := range cm.config.Connections {
		if c.ID != id {
			filtered = append(filtered, c)
		}
	}
	cm.config.Connections = filtered

	if cm.config.ActiveConnectionID == id {
		if len(cm.config.Connections) > 0 {
			cm.config.ActiveConnectionID = cm.config.Connections[0].ID
		} else {
			cm.config.ActiveConnectionID = ""
		}
	}

	return cm.save()
}

// SetActiveConnection sets the active connection ID.
func (cm *ConfigManager) SetActiveConnection(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	previous := cm.config.ActiveConnectionID
	cm.config.ActiveConnectionID = id
	if err := cm.save(); err != nil {
		cm.config.ActiveConnectionID = previous
		return err
	}
	return nil
}

// GetActiveConnection returns the currently selected connection profile with its decrypted password from OS Keychain.
func (cm *ConfigManager) GetActiveConnection() *models.ConnectionProfile {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var target *models.ConnectionProfile
	for _, c := range cm.config.Connections {
		if c.ID == cm.config.ActiveConnectionID {
			cpy := c
			target = &cpy
			break
		}
	}
	if target == nil && len(cm.config.Connections) > 0 {
		cpy := cm.config.Connections[0]
		target = &cpy
	}

	if target != nil {
		// Populate password from OS Keychain
		pass, _ := cm.secretMgr.GetPassword(target.ID)
		target.Password = pass
	}

	return target
}

// GetConnectionWithSecret returns a specific connection profile with password populated from OS Keychain.
func (cm *ConfigManager) GetConnectionWithSecret(id string) *models.ConnectionProfile {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, c := range cm.config.Connections {
		if c.ID == id {
			cpy := c
			pass, _ := cm.secretMgr.GetPassword(c.ID)
			cpy.Password = pass
			return &cpy
		}
	}
	return nil
}

// GetSecurityPolicy returns the active security policy.
func (cm *ConfigManager) GetSecurityPolicy() models.SecurityPolicy {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.SecurityPolicy
}

// SaveSecurityPolicy updates and persists the security policy.
func (cm *ConfigManager) SaveSecurityPolicy(policy models.SecurityPolicy) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.SecurityPolicy = policy
	return cm.save()
}

// GetMCPPort returns the configured MCP HTTP/SSE server port.
func (cm *ConfigManager) GetMCPPort() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.config.MCPPort <= 0 {
		return 8085
	}
	return cm.config.MCPPort
}

// SetMCPPort updates the configured port.
func (cm *ConfigManager) SetMCPPort(port int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.MCPPort = port
	return cm.save()
}
