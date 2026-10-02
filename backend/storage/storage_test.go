package storage

import (
	"path/filepath"
	"testing"
)

func TestSetActiveConnectionRollsBackWhenSaveFails(t *testing.T) {
	cm := &ConfigManager{filePath: filepath.Join(t.TempDir(), "missing", "config.json"), config: AppConfig{ActiveConnectionID: "original"}}
	if err := cm.SetActiveConnection("new"); err == nil {
		t.Fatal("expected persistence failure")
	}
	if got := cm.GetConfig().ActiveConnectionID; got != "original" {
		t.Fatalf("active connection changed despite failure: %s", got)
	}
}
