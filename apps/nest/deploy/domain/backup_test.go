package domain

import (
	"fmt"
	"testing"
)

func TestBackup_NoConflict(t *testing.T) {
	exists := func(name string) bool { return false }
	result := NextBackupName("foo", ".bak", exists)
	if result != "foo.bak" {
		t.Errorf("expected 'foo.bak', got '%s'", result)
	}
}

func TestBackup_FirstConflict(t *testing.T) {
	taken := map[string]bool{"foo.bak": true}
	exists := func(name string) bool { return taken[name] }
	result := NextBackupName("foo", ".bak", exists)
	if result != "foo.bak.2" {
		t.Errorf("expected 'foo.bak.2', got '%s'", result)
	}
}

func TestBackup_MultiConflict(t *testing.T) {
	taken := map[string]bool{"foo.bak": true, "foo.bak.2": true}
	exists := func(name string) bool { return taken[name] }
	result := NextBackupName("foo", ".bak", exists)
	if result != "foo.bak.3" {
		t.Errorf("expected 'foo.bak.3', got '%s'", result)
	}
}

func TestBackup_CustomSuffix(t *testing.T) {
	exists := func(name string) bool { return false }
	result := NextBackupName("foo", ".old", exists)
	if result != "foo.old" {
		t.Errorf("expected 'foo.old', got '%s'", result)
	}
}

func TestBackup_HighSequence(t *testing.T) {
	taken := make(map[string]bool)
	taken["data.bak"] = true
	for i := 2; i <= 10; i++ {
		taken[fmt.Sprintf("data.bak.%d", i)] = true
	}
	exists := func(name string) bool { return taken[name] }
	result := NextBackupName("data", ".bak", exists)
	if result != "data.bak.11" {
		t.Errorf("expected 'data.bak.11', got '%s'", result)
	}
}

func TestBackup_GapInSequence(t *testing.T) {
	// .bak and .bak.3 exist, but .bak.2 does not → should pick .bak.2
	taken := map[string]bool{"config.bak": true, "config.bak.3": true}
	exists := func(name string) bool { return taken[name] }
	result := NextBackupName("config", ".bak", exists)
	if result != "config.bak.2" {
		t.Errorf("expected 'config.bak.2', got '%s'", result)
	}
}
