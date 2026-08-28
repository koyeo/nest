package domain

import (
	"testing"
	"time"
)

func TestIsManaged_FileInEntry(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "app.js"}, {Path: "index.html"}}},
		},
	}
	if !s.IsManaged("app.js") {
		t.Error("expected app.js to be managed")
	}
}

func TestIsManaged_FileNotInEntry(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "app.js"}}},
		},
	}
	if s.IsManaged("other.js") {
		t.Error("expected other.js to NOT be managed")
	}
}

func TestIsManaged_NilSnapshot(t *testing.T) {
	var s *Snapshot
	if s.IsManaged("anything") {
		t.Error("expected nil snapshot to return false")
	}
}

func TestIsManaged_MultipleEntries(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "old.js"}}},
			{Files: []FileRecord{{Path: "new.js"}}},
		},
	}
	if !s.IsManaged("old.js") {
		t.Error("expected old.js from historical entry to be managed")
	}
	if !s.IsManaged("new.js") {
		t.Error("expected new.js from latest entry to be managed")
	}
}

func TestAddEntry(t *testing.T) {
	s := &Snapshot{}
	if len(s.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(s.Entries))
	}

	entry := NewSnapshotEntry("bundle.tar.gz", "abc123", []FileRecord{
		{Path: "app.js", Hash: "def456", ModTime: time.Now()},
	})
	s.AddEntry(entry)

	if len(s.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(s.Entries))
	}
	if s.Entries[0].BundleName != "bundle.tar.gz" {
		t.Errorf("expected bundle name 'bundle.tar.gz', got '%s'", s.Entries[0].BundleName)
	}
	if len(s.Entries[0].Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(s.Entries[0].Files))
	}
}
