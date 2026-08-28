package domain

// ConflictAction represents the user's choice for handling a file conflict.
type ConflictAction int

const (
	// ActionBackup renames the conflicting file with a suffix.
	ActionBackup ConflictAction = iota
	// ActionRemove deletes the conflicting file.
	ActionRemove
	// ActionOverwrite removes conflicting files without backup (non-interactive).
	ActionOverwrite
)

// ConflictResult separates conflicting files into managed (Nest-tracked)
// and unmanaged (external) categories.
type ConflictResult struct {
	// ManagedFiles are files tracked in the snapshot — safe to replace directly.
	ManagedFiles []string
	// UnmanagedFiles are files not in the snapshot — require user decision.
	UnmanagedFiles []string
}

// ClassifyConflicts is a pure function that categorizes conflicting files
// based on whether they are tracked in the snapshot.
func ClassifyConflicts(conflicts []string, snapshot *Snapshot) ConflictResult {
	result := ConflictResult{}
	for _, f := range conflicts {
		if snapshot != nil && snapshot.IsManaged(f) {
			result.ManagedFiles = append(result.ManagedFiles, f)
		} else {
			result.UnmanagedFiles = append(result.UnmanagedFiles, f)
		}
	}
	return result
}
