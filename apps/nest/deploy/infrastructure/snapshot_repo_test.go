package infrastructure

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/koyeo/nest/deploy/domain"
)

// --- Mock RemoteFS for snapshot repo tests ---

type mockFS struct {
	files map[string][]byte
	dirs  map[string]bool
}

func newMockFS() *mockFS {
	return &mockFS{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

type mockInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (m *mockInfo) Name() string      { return m.name }
func (m *mockInfo) Size() int64       { return m.size }
func (m *mockInfo) Mode() fs.FileMode { return 0644 }
func (m *mockInfo) ModTime() time.Time { return m.modTime }
func (m *mockInfo) IsDir() bool       { return m.isDir }

func (m *mockFS) Stat(path string) (domain.FileInfo, error) {
	if _, ok := m.files[path]; ok {
		return &mockInfo{name: path, size: int64(len(m.files[path]))}, nil
	}
	if m.dirs[path] {
		return &mockInfo{name: path, isDir: true}, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockFS) MkdirAll(path string) error {
	m.dirs[path] = true
	return nil
}

func (m *mockFS) ReadDir(path string) ([]string, error) { return nil, nil }

func (m *mockFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return data, nil
}

func (m *mockFS) WriteFile(path string, data []byte) error {
	m.files[path] = data
	return nil
}

func (m *mockFS) Remove(path string) error {
	delete(m.files, path)
	return nil
}

func (m *mockFS) Rename(src, dst string) error {
	m.files[dst] = m.files[src]
	delete(m.files, src)
	return nil
}

func (m *mockFS) FileHash(path string) (string, error)           { return "mock", nil }
func (m *mockFS) FileModTime(path string) (time.Time, error)     { return time.Now(), nil }

// --- Tests ---

func TestRead_NoFile(t *testing.T) {
	mockfs := newMockFS()
	repo := NewSnapshotRepo(mockfs)

	snap, err := repo.Read("/target")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if snap != nil {
		t.Error("expected nil snapshot when file doesn't exist")
	}
}

func TestRead_ValidJSON(t *testing.T) {
	mockfs := newMockFS()
	snap := &domain.Snapshot{
		Entries: []domain.SnapshotEntry{
			{BundleName: "app.tar.gz", BundleHash: "abc123"},
		},
	}
	data, _ := json.Marshal(snap)
	mockfs.files["/target/.nest/snapshot.json"] = data

	repo := NewSnapshotRepo(mockfs)
	result, err := repo.Read("/target")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if result == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].BundleName != "app.tar.gz" {
		t.Errorf("expected bundle name 'app.tar.gz', got '%s'", result.Entries[0].BundleName)
	}
}

func TestRead_InvalidJSON(t *testing.T) {
	mockfs := newMockFS()
	mockfs.files["/target/.nest/snapshot.json"] = []byte("{invalid json")

	repo := NewSnapshotRepo(mockfs)
	_, err := repo.Read("/target")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestWrite_CreatesDir(t *testing.T) {
	mockfs := newMockFS()
	repo := NewSnapshotRepo(mockfs)

	snap := &domain.Snapshot{}
	err := repo.Write("/target", snap)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !mockfs.dirs["/target/.nest"] {
		t.Error("expected .nest directory to be created")
	}
}

func TestWrite_RoundTrip(t *testing.T) {
	mockfs := newMockFS()
	repo := NewSnapshotRepo(mockfs)

	original := &domain.Snapshot{
		Entries: []domain.SnapshotEntry{
			{
				BundleName: "test.tar.gz",
				BundleHash: "deadbeef",
				Files: []domain.FileRecord{
					{Path: "main.go", Hash: "abc123"},
				},
			},
		},
	}

	err := repo.Write("/target", original)
	if err != nil {
		t.Fatalf("write error: %s", err)
	}

	result, err := repo.Read("/target")
	if err != nil {
		t.Fatalf("read error: %s", err)
	}
	if result == nil {
		t.Fatal("expected non-nil snapshot after round trip")
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].BundleName != "test.tar.gz" {
		t.Errorf("bundle name mismatch: '%s'", result.Entries[0].BundleName)
	}
	if result.Entries[0].Files[0].Path != "main.go" {
		t.Errorf("file path mismatch: '%s'", result.Entries[0].Files[0].Path)
	}
}
