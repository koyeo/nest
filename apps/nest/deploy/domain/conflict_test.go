package domain

import "testing"

func TestClassify_NoSnapshot(t *testing.T) {
	result := ClassifyConflicts([]string{"a.js", "b.js"}, nil)
	if len(result.ManagedFiles) != 0 {
		t.Errorf("expected 0 managed, got %d", len(result.ManagedFiles))
	}
	if len(result.UnmanagedFiles) != 2 {
		t.Errorf("expected 2 unmanaged, got %d", len(result.UnmanagedFiles))
	}
}

func TestClassify_AllManaged(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "a.js"}, {Path: "b.js"}}},
		},
	}
	result := ClassifyConflicts([]string{"a.js", "b.js"}, s)
	if len(result.ManagedFiles) != 2 {
		t.Errorf("expected 2 managed, got %d", len(result.ManagedFiles))
	}
	if len(result.UnmanagedFiles) != 0 {
		t.Errorf("expected 0 unmanaged, got %d", len(result.UnmanagedFiles))
	}
}

func TestClassify_AllUnmanaged(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "x.js"}}},
		},
	}
	result := ClassifyConflicts([]string{"a.js", "b.js"}, s)
	if len(result.ManagedFiles) != 0 {
		t.Errorf("expected 0 managed, got %d", len(result.ManagedFiles))
	}
	if len(result.UnmanagedFiles) != 2 {
		t.Errorf("expected 2 unmanaged, got %d", len(result.UnmanagedFiles))
	}
}

func TestClassify_Mixed(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "a.js"}}},
		},
	}
	result := ClassifyConflicts([]string{"a.js", "b.js", "c.js"}, s)
	if len(result.ManagedFiles) != 1 {
		t.Errorf("expected 1 managed, got %d", len(result.ManagedFiles))
	}
	if result.ManagedFiles[0] != "a.js" {
		t.Errorf("expected managed file 'a.js', got '%s'", result.ManagedFiles[0])
	}
	if len(result.UnmanagedFiles) != 2 {
		t.Errorf("expected 2 unmanaged, got %d", len(result.UnmanagedFiles))
	}
}

func TestClassify_EmptyConflicts(t *testing.T) {
	s := &Snapshot{
		Entries: []SnapshotEntry{
			{Files: []FileRecord{{Path: "a.js"}}},
		},
	}
	result := ClassifyConflicts([]string{}, s)
	if len(result.ManagedFiles) != 0 {
		t.Errorf("expected 0 managed, got %d", len(result.ManagedFiles))
	}
	if len(result.UnmanagedFiles) != 0 {
		t.Errorf("expected 0 unmanaged, got %d", len(result.UnmanagedFiles))
	}
}
