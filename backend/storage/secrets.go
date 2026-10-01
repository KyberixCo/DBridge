package storage

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const ServiceName = "kyberix-dbridge"

// SecretManager handles native OS credential vault storage (macOS Keychain, Windows Credential Manager).
type SecretManager struct {
	service string
}

// NewSecretManager creates a SecretManager instance.
func NewSecretManager() *SecretManager {
	return &SecretManager{
		service: ServiceName,
	}
}

// SetPassword securely stores the database password in the OS Keychain/Credential Manager.
func (sm *SecretManager) SetPassword(connectionID string, password string) error {
	if connectionID == "" {
		return errors.New("connectionID cannot be empty")
	}
	if password == "" {
		// If password is empty, remove any existing key
		_ = keyring.Delete(sm.service, connectionID)
		return nil
	}
	if err := keyring.Set(sm.service, connectionID, password); err != nil {
		return fmt.Errorf("failed to save password in OS keychain: %w", err)
	}
	return nil
}

// GetPassword retrieves the database password from the OS Keychain/Credential Manager.
func (sm *SecretManager) GetPassword(connectionID string) (string, error) {
	if connectionID == "" {
		return "", errors.New("connectionID cannot be empty")
	}
	pass, err := keyring.Get(sm.service, connectionID)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil // no password saved
		}
		return "", fmt.Errorf("failed to retrieve password from OS keychain: %w", err)
	}
	return pass, nil
}

// DeletePassword removes a stored credential from the OS Keychain/Credential Manager.
func (sm *SecretManager) DeletePassword(connectionID string) error {
	if connectionID == "" {
		return nil
	}
	err := keyring.Delete(sm.service, connectionID)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("failed to delete password from OS keychain: %w", err)
	}
	return nil
}
