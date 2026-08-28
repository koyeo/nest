package domain

import (
	"io/fs"
	"time"
)

// FileInfo represents basic file metadata from a remote filesystem.
type FileInfo interface {
	Name() string
	Size() int64
	Mode() fs.FileMode
	ModTime() time.Time
	IsDir() bool
}

// RemoteFS abstracts remote filesystem operations (e.g., SFTP).
type RemoteFS interface {
	// Stat returns file info. Returns error if the file does not exist.
	Stat(path string) (FileInfo, error)
	// MkdirAll creates a directory and all parents.
	MkdirAll(path string) error
	// ReadDir returns child names in a directory (non-recursive).
	ReadDir(path string) ([]string, error)
	// ReadFile reads file contents.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes content to a file, creating it if necessary.
	WriteFile(path string, data []byte) error
	// Remove removes a file or directory recursively.
	Remove(path string) error
	// Rename moves/renames a file or directory.
	Rename(src, dst string) error
	// FileHash computes SHA256 hash of a remote file.
	FileHash(path string) (string, error)
	// FileModTime returns the modification time of a file.
	FileModTime(path string) (time.Time, error)
}

// RemoteExec abstracts remote command execution (e.g., SSH).
type RemoteExec interface {
	// Exec runs a command and returns combined output error.
	Exec(command string) error
	// ExecPipe runs a command with stdout/stderr piped to the console.
	ExecPipe(command string) error
}

// UserPrompter abstracts user interaction for conflict resolution.
type UserPrompter interface {
	// AskConflictAction asks the user how to handle conflicting files.
	// Returns the chosen action and backup suffix (if backup was chosen).
	AskConflictAction(files []string, lang string) (ConflictAction, string, error)
}

// SnapshotRepository abstracts reading/writing snapshot metadata.
type SnapshotRepository interface {
	// Read returns the snapshot from the target directory.
	// Returns nil (no error) if no snapshot exists.
	Read(targetDir string) (*Snapshot, error)
	// Write persists the snapshot to the target directory.
	Write(targetDir string, snapshot *Snapshot) error
}
