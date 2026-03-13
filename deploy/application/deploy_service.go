package application

import (
	"fmt"
	"time"

	"github.com/koyeo/nest/deploy/domain"
	"github.com/koyeo/nest/i18n"
)

// DeployService orchestrates the deploy conflict resolution flow.
type DeployService struct {
	fs       domain.RemoteFS
	exec     domain.RemoteExec
	snapshot domain.SnapshotRepository
	prompter domain.UserPrompter
	lang     string
}

// NewDeployService creates a new DeployService with all dependencies injected.
func NewDeployService(
	fs domain.RemoteFS,
	exec domain.RemoteExec,
	snapshot domain.SnapshotRepository,
	prompter domain.UserPrompter,
	lang string,
) *DeployService {
	return &DeployService{
		fs:       fs,
		exec:     exec,
		snapshot: snapshot,
		prompter: prompter,
		lang:     lang,
	}
}

// Deploy handles extracting, conflict resolution, file deployment, and snapshot update.
//
// Parameters:
//   - bundleRemotePath: path to the uploaded tar.gz on the remote server
//   - targetDir: the target directory where files should be deployed
//   - bundleName: name of the bundle (e.g., "app.tar.gz")
//   - bundleHash: SHA256 hash of the bundle
func (s *DeployService) Deploy(bundleRemotePath, targetDir, bundleName, bundleHash, conflictStrategy string) error {
	nestDir := fmt.Sprintf("%s/.nest", targetDir)
	tmpDir := fmt.Sprintf("%s/.nest/tmp", targetDir)

	// Ensure .nest directory exists
	_ = s.fs.MkdirAll(nestDir)

	// Extract to .nest/tmp/
	cmd := fmt.Sprintf("rm -rf %s && mkdir -p %s && tar -xzf %s -C %s", tmpDir, tmpDir, bundleRemotePath, tmpDir)
	if err := s.exec.Exec(cmd); err != nil {
		return fmt.Errorf("extract bundle error: %s", err)
	}
	defer func() {
		// Clean up .nest/tmp/
		_ = s.exec.Exec(fmt.Sprintf("rm -rf %s", tmpDir))
	}()

	// List extracted files
	extractedFiles, err := s.fs.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("list extracted files error: %s", err)
	}

	// Detect conflicts
	var conflictFiles []string
	for _, f := range extractedFiles {
		filePath := fmt.Sprintf("%s/%s", targetDir, f)
		if _, statErr := s.fs.Stat(filePath); statErr == nil {
			conflictFiles = append(conflictFiles, f)
		}
	}

	// Read existing snapshot
	snap, err := s.snapshot.Read(targetDir)
	if err != nil {
		return fmt.Errorf("read snapshot error: %s", err)
	}

	// Resolve conflicts
	if len(conflictFiles) > 0 {
		if err = s.resolveConflicts(targetDir, conflictFiles, snap, conflictStrategy); err != nil {
			return err
		}
	}

	// Move files from .nest/tmp/ to target directory
	for _, f := range extractedFiles {
		src := fmt.Sprintf("%s/%s", tmpDir, f)
		dst := fmt.Sprintf("%s/%s", targetDir, f)
		cmd = fmt.Sprintf("mv %s %s", src, dst)
		if err = s.exec.Exec(cmd); err != nil {
			return fmt.Errorf("move file error: %s", err)
		}
	}

	// Build file records for snapshot
	var fileRecords []domain.FileRecord
	for _, f := range extractedFiles {
		filePath := fmt.Sprintf("%s/%s", targetDir, f)
		hash, _ := s.fs.FileHash(filePath)
		modTime, mtErr := s.fs.FileModTime(filePath)
		if mtErr != nil {
			modTime = time.Now().UTC()
		}
		fileRecords = append(fileRecords, domain.FileRecord{
			Path:    f,
			Hash:    hash,
			ModTime: modTime,
		})
	}

	// Update snapshot
	entry := domain.NewSnapshotEntry(bundleName, bundleHash, fileRecords)
	isNew := snap == nil
	if isNew {
		snap = &domain.Snapshot{}
	}
	snap.AddEntry(entry)

	if err = s.snapshot.Write(targetDir, snap); err != nil {
		return fmt.Errorf("write snapshot error: %s", err)
	}

	fmt.Println(i18n.Msg(i18n.MsgDeployComplete, s.lang))
	return nil
}

// resolveConflicts handles managed and unmanaged file conflicts.
func (s *DeployService) resolveConflicts(targetDir string, conflictFiles []string, snap *domain.Snapshot, conflictStrategy string) error {
	result := domain.ClassifyConflicts(conflictFiles, snap)

	// Remove Nest-managed files silently
	for _, f := range result.ManagedFiles {
		if err := s.fs.Remove(fmt.Sprintf("%s/%s", targetDir, f)); err != nil {
			return fmt.Errorf("remove managed file error: %s", err)
		}
	}

	// Handle unmanaged files
	if len(result.UnmanagedFiles) > 0 {
		var action domain.ConflictAction
		var suffix string
		var err error

		switch conflictStrategy {
		case "overwrite":
			action = domain.ActionOverwrite
		case "backup":
			action = domain.ActionBackup
			suffix = ".bak"
		case "error":
			return fmt.Errorf("unmanaged file conflicts: %v", result.UnmanagedFiles)
		default:
			// Interactive mode
			action, suffix, err = s.prompter.AskConflictAction(result.UnmanagedFiles, s.lang)
			if err != nil {
				return err
			}
		}

		for _, f := range result.UnmanagedFiles {
			filePath := fmt.Sprintf("%s/%s", targetDir, f)
			switch action {
			case domain.ActionBackup:
				backupName := domain.NextBackupName(f, suffix, func(candidate string) bool {
					_, statErr := s.fs.Stat(fmt.Sprintf("%s/%s", targetDir, candidate))
					return statErr == nil
				})
				fmt.Println(i18n.Msgf(i18n.MsgBackingUp, s.lang, f, backupName))
				if err = s.fs.Rename(filePath, fmt.Sprintf("%s/%s", targetDir, backupName)); err != nil {
					return fmt.Errorf("backup file error: %s", err)
				}
			case domain.ActionRemove, domain.ActionOverwrite:
				if action == domain.ActionRemove {
					fmt.Println(i18n.Msgf(i18n.MsgRemoving, s.lang, f))
				}
				if err = s.fs.Remove(filePath); err != nil {
					return fmt.Errorf("remove file error: %s", err)
				}
			}
		}
	}

	return nil
}

