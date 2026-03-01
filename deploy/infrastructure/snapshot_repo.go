package infrastructure

import (
	"encoding/json"
	"fmt"

	"github.com/koyeo/nest/deploy/domain"
)

const (
	nestMetaDir      = ".nest"
	snapshotFileName = ".nest/snapshot.json"
)

// SnapshotRepo implements domain.SnapshotRepository using RemoteFS.
type SnapshotRepo struct {
	fs domain.RemoteFS
}

// NewSnapshotRepo creates a new SnapshotRepo.
func NewSnapshotRepo(fs domain.RemoteFS) *SnapshotRepo {
	return &SnapshotRepo{fs: fs}
}

func (r *SnapshotRepo) Read(targetDir string) (*domain.Snapshot, error) {
	snapshotPath := fmt.Sprintf("%s/%s", targetDir, snapshotFileName)

	// Check if file exists
	_, err := r.fs.Stat(snapshotPath)
	if err != nil {
		// File does not exist
		return nil, nil
	}

	data, err := r.fs.ReadFile(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("read snapshot error: %s", err)
	}

	var snapshot domain.Snapshot
	if err = json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decode snapshot error: %s", err)
	}
	return &snapshot, nil
}

func (r *SnapshotRepo) Write(targetDir string, snapshot *domain.Snapshot) error {
	metaDir := fmt.Sprintf("%s/%s", targetDir, nestMetaDir)

	// Ensure .nest directory exists
	if err := r.fs.MkdirAll(metaDir); err != nil {
		return fmt.Errorf("create .nest dir error: %s", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot error: %s", err)
	}

	snapshotPath := fmt.Sprintf("%s/%s", targetDir, snapshotFileName)
	if err = r.fs.WriteFile(snapshotPath, data); err != nil {
		return fmt.Errorf("write snapshot error: %s", err)
	}
	return nil
}
