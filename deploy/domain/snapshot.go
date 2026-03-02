package domain

import "time"

// Snapshot is the aggregate root representing deployment metadata
// stored at .nest/snapshot.json on the remote server.
type Snapshot struct {
	Entries []SnapshotEntry `json:"entries"`
}

// SnapshotEntry records a single deployment event.
type SnapshotEntry struct {
	BundleName string       `json:"bundle_name"`
	BundleHash string       `json:"bundle_hash"`
	DeployedAt time.Time    `json:"deployed_at"`
	Files      []FileRecord `json:"files"`
}

// FileRecord is a value object storing metadata for a deployed file.
type FileRecord struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	ModTime time.Time `json:"mod_time"`
}

// IsManaged checks if a filename appears in any entry of the snapshot.
func (s *Snapshot) IsManaged(filename string) bool {
	if s == nil {
		return false
	}
	for _, entry := range s.Entries {
		for _, f := range entry.Files {
			if f.Path == filename {
				return true
			}
		}
	}
	return false
}

// AddEntry upserts a deployment entry by BundleName.
// If an entry with the same BundleName exists, it is replaced; otherwise appended.
func (s *Snapshot) AddEntry(entry SnapshotEntry) {
	for i, e := range s.Entries {
		if e.BundleName == entry.BundleName {
			s.Entries[i] = entry
			return
		}
	}
	s.Entries = append(s.Entries, entry)
}

// NewSnapshotEntry creates a new SnapshotEntry with the current time.
func NewSnapshotEntry(bundleName, bundleHash string, files []FileRecord) SnapshotEntry {
	return SnapshotEntry{
		BundleName: bundleName,
		BundleHash: bundleHash,
		DeployedAt: time.Now().UTC(),
		Files:      files,
	}
}
