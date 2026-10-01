package storage

import (
	"testing"
)

func TestSecretManager_Lifecycle(t *testing.T) {
	sm := NewSecretManager()
	testConnID := "test-conn-lifecycle-12345"
	secretPass := "UltraSecretPassword99!"

	// 1. Set password
	if err := sm.SetPassword(testConnID, secretPass); err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}

	// 2. Retrieve password
	gotPass, err := sm.GetPassword(testConnID)
	if err != nil {
		t.Fatalf("GetPassword failed: %v", err)
	}
	if gotPass != secretPass {
		t.Errorf("Expected %q, got %q", secretPass, gotPass)
	}

	// 3. Delete password
	if err := sm.DeletePassword(testConnID); err != nil {
		t.Fatalf("DeletePassword failed: %v", err)
	}

	// 4. Verify deleted
	afterPass, err := sm.GetPassword(testConnID)
	if err != nil {
		t.Fatalf("GetPassword after delete returned error: %v", err)
	}
	if afterPass != "" {
		t.Errorf("Expected empty password after delete, got %q", afterPass)
	}
}
