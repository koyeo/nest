package application

import (
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/koyeo/nest/deploy/domain"
)

// --- Mocks ---

type mockRemoteFS struct {
	files map[string][]byte // path → content
	dirs  map[string]bool   // directories
}

func newMockFS() *mockRemoteFS {
	return &mockRemoteFS{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

type mockFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string      { return m.name }
func (m *mockFileInfo) Size() int64       { return m.size }
func (m *mockFileInfo) Mode() fs.FileMode { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool       { return m.isDir }

func (m *mockRemoteFS) Stat(path string) (domain.FileInfo, error) {
	if _, ok := m.files[path]; ok {
		return &mockFileInfo{name: path, size: int64(len(m.files[path])), modTime: time.Now()}, nil
	}
	if m.dirs[path] {
		return &mockFileInfo{name: path, isDir: true, modTime: time.Now()}, nil
	}
	return nil, fmt.Errorf("not found: %s", path)
}

func (m *mockRemoteFS) MkdirAll(path string) error {
	m.dirs[path] = true
	return nil
}

func (m *mockRemoteFS) ReadDir(path string) ([]string, error) {
	var names []string
	prefix := path + "/"
	for p := range m.files {
		if len(p) > len(prefix) && p[:len(prefix)] == prefix {
			rest := p[len(prefix):]
			// Only direct children
			if idx := indexOf(rest, '/'); idx == -1 {
				names = append(names, rest)
			}
		}
	}
	return names, nil
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func (m *mockRemoteFS) ReadFile(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("not found: %s", path)
	}
	return data, nil
}

func (m *mockRemoteFS) WriteFile(path string, data []byte) error {
	m.files[path] = data
	return nil
}

func (m *mockRemoteFS) Remove(path string) error {
	delete(m.files, path)
	delete(m.dirs, path)
	return nil
}

func (m *mockRemoteFS) Rename(src, dst string) error {
	if data, ok := m.files[src]; ok {
		m.files[dst] = data
		delete(m.files, src)
		return nil
	}
	return fmt.Errorf("not found: %s", src)
}

func (m *mockRemoteFS) FileHash(path string) (string, error) {
	return "mockhash", nil
}

func (m *mockRemoteFS) FileModTime(path string) (time.Time, error) {
	return time.Now().UTC(), nil
}

type mockRemoteExec struct {
	commands []string
}

func newMockExec() *mockRemoteExec {
	return &mockRemoteExec{}
}

func (m *mockRemoteExec) Exec(command string) error {
	m.commands = append(m.commands, command)
	return nil
}

func (m *mockRemoteExec) ExecPipe(command string) error {
	m.commands = append(m.commands, command)
	return nil
}

type mockSnapshotRepo struct {
	snapshots map[string]*domain.Snapshot
}

func newMockSnapshotRepo() *mockSnapshotRepo {
	return &mockSnapshotRepo{snapshots: make(map[string]*domain.Snapshot)}
}

func (m *mockSnapshotRepo) Read(targetDir string) (*domain.Snapshot, error) {
	s, ok := m.snapshots[targetDir]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockSnapshotRepo) Write(targetDir string, snapshot *domain.Snapshot) error {
	m.snapshots[targetDir] = snapshot
	return nil
}

type mockPrompter struct {
	action domain.ConflictAction
	suffix string
}

func (m *mockPrompter) AskConflictAction(files []string, lang string) (domain.ConflictAction, string, error) {
	return m.action, m.suffix, nil
}

// --- Helper ---

func setupService(
	mockFS *mockRemoteFS,
	mockExec *mockRemoteExec,
	mockRepo *mockSnapshotRepo,
	prompter domain.UserPrompter,
) *DeployService {
	return NewDeployService(mockFS, mockExec, mockRepo, prompter, "en")
}

// --- Tests ---

func TestDeploy_EmptyDir(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionBackup, suffix: ".bak"}

	// Simulate extracted files in .nest/tmp/
	mockFS.files["/target/.nest/tmp/app.js"] = []byte("content")

	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	err := svc.Deploy("/target/bundle.tar.gz", "/target", "app.tar.gz", "hash123", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Snapshot should be created
	snap := mockRepo.snapshots["/target"]
	if snap == nil {
		t.Fatal("expected snapshot to be created")
	}
	if len(snap.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(snap.Entries))
	}
}

func TestDeploy_HasSnapshotManagedFiles(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionBackup, suffix: ".bak"}

	// Existing managed file
	mockFS.files["/target/app.js"] = []byte("old")
	mockRepo.snapshots["/target"] = &domain.Snapshot{
		Entries: []domain.SnapshotEntry{
			{Files: []domain.FileRecord{{Path: "app.js"}}},
		},
	}

	// New file to deploy
	mockFS.files["/target/.nest/tmp/app.js"] = []byte("new")

	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	err := svc.Deploy("/target/bundle.tar.gz", "/target", "app.tar.gz", "hash123", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Old file should be removed (managed files get silently replaced)
	if _, err := mockFS.Stat("/target/app.js.bak"); err == nil {
		t.Error("managed file should NOT be backed up")
	}
}

func TestDeploy_HasSnapshotUnmanagedFiles(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionBackup, suffix: ".bak"}

	// Existing unmanaged file
	mockFS.files["/target/config.yml"] = []byte("secret")
	mockRepo.snapshots["/target"] = &domain.Snapshot{
		Entries: []domain.SnapshotEntry{
			{Files: []domain.FileRecord{{Path: "app.js"}}},
		},
	}

	// New file to deploy with same name
	mockFS.files["/target/.nest/tmp/config.yml"] = []byte("new config")

	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	err := svc.Deploy("/target/bundle.tar.gz", "/target", "app.tar.gz", "hash123", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Unmanaged file should be backed up
	if _, err := mockFS.Stat("/target/config.yml.bak"); err != nil {
		t.Error("expected unmanaged file to be backed up as config.yml.bak")
	}
}

func TestDeploy_NoSnapshotWithConflict(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionBackup, suffix: ".bak"}

	// Existing file, no snapshot
	mockFS.files["/target/app.js"] = []byte("old")
	mockFS.files["/target/.nest/tmp/app.js"] = []byte("new")

	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	err := svc.Deploy("/target/bundle.tar.gz", "/target", "app.tar.gz", "hash123", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// File should be backed up (no snapshot → unmanaged)
	if _, err := mockFS.Stat("/target/app.js.bak"); err != nil {
		t.Error("expected file to be backed up when no snapshot exists")
	}
}

func TestDeploy_UserChoosesRemove(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionRemove, suffix: ""}

	// Existing file, no snapshot
	mockFS.files["/target/old.txt"] = []byte("data")
	mockFS.files["/target/.nest/tmp/old.txt"] = []byte("new data")

	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	err := svc.Deploy("/target/bundle.tar.gz", "/target", "app.tar.gz", "hash123", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// No backup should exist
	if _, err := mockFS.Stat("/target/old.txt.bak"); err == nil {
		t.Error("expected NO backup when user chose remove")
	}
}

func TestDeploy_SnapshotWrittenAfterDeploy(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionBackup, suffix: ".bak"}

	mockFS.files["/target/.nest/tmp/index.html"] = []byte("<html>")

	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	err := svc.Deploy("/target/bundle.tar.gz", "/target", "bundle.tar.gz", "hash999", "")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	snap := mockRepo.snapshots["/target"]
	if snap == nil {
		t.Fatal("snapshot not written")
	}
	if snap.Entries[0].BundleHash != "hash999" {
		t.Errorf("expected hash 'hash999', got '%s'", snap.Entries[0].BundleHash)
	}
	if len(snap.Entries[0].Files) != 1 || snap.Entries[0].Files[0].Path != "index.html" {
		t.Errorf("snapshot file record mismatch")
	}
}

func TestDeploy_SnapshotPreservesHistory(t *testing.T) {
	mockFS := newMockFS()
	mockExec := newMockExec()
	mockRepo := newMockSnapshotRepo()
	prompter := &mockPrompter{action: domain.ActionBackup, suffix: ".bak"}

	// First deploy
	mockFS.files["/target/.nest/tmp/v1.js"] = []byte("v1")
	svc := setupService(mockFS, mockExec, mockRepo, prompter)
	_ = svc.Deploy("/target/bundle.tar.gz", "/target", "v1.tar.gz", "hash1", "")

	// Second deploy — existing file is now managed
	mockFS.files["/target/.nest/tmp/v2.js"] = []byte("v2")
	_ = svc.Deploy("/target/bundle2.tar.gz", "/target", "v2.tar.gz", "hash2", "")

	snap := mockRepo.snapshots["/target"]
	if snap == nil {
		t.Fatal("snapshot not written")
	}
	if len(snap.Entries) != 2 {
		t.Errorf("expected 2 entries (history preserved), got %d", len(snap.Entries))
	}
}
