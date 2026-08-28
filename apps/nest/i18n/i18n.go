package i18n

import "fmt"

// Message keys
const (
	MsgConflictFound   = "conflict_found"
	MsgChooseAction    = "choose_action"
	MsgBackupSuffix    = "backup_suffix"
	MsgBackingUp       = "backing_up"
	MsgRemoving        = "removing"
	MsgDeployComplete  = "deploy_complete"
	MsgSnapshotCreated = "snapshot_created"
	MsgSnapshotUpdated = "snapshot_updated"
)

var messages = map[string]map[string]string{
	MsgConflictFound: {
		"zh": "⚠ Unmanaged file conflicts in target directory: %s",
		"en": "⚠ Unmanaged file conflicts in target directory: %s",
	},
	MsgChooseAction: {
		"zh": "  [1] Backup (default)\n  [2] Remove\nChoose [1]: ",
		"en": "  [1] Backup (default)\n  [2] Remove\nChoose [1]: ",
	},
	MsgBackupSuffix: {
		"zh": "Backup suffix [.bak]: ",
		"en": "Backup suffix [.bak]: ",
	},
	MsgBackingUp: {
		"zh": "  📦 Backup: %s → %s",
		"en": "  📦 Backup: %s → %s",
	},
	MsgRemoving: {
		"zh": "  🗑  Remove: %s",
		"en": "  🗑  Remove: %s",
	},
	MsgDeployComplete: {
		"zh": "  ✅ Deploy complete",
		"en": "  ✅ Deploy complete",
	},
	MsgSnapshotCreated: {
		"zh": "  📝 Created snapshot: .nest/snapshot.json",
		"en": "  📝 Created snapshot: .nest/snapshot.json",
	},
	MsgSnapshotUpdated: {
		"zh": "  📝 Updated snapshot: .nest/snapshot.json",
		"en": "  📝 Updated snapshot: .nest/snapshot.json",
	},
}

// Msg returns a localized message by key and language code.
func Msg(key, lang string) string {
	m, ok := messages[key]
	if !ok {
		return key
	}
	s, ok := m[lang]
	if !ok {
		s = m["en"]
	}
	return s
}

// Msgf returns a formatted localized message.
func Msgf(key, lang string, args ...interface{}) string {
	return fmt.Sprintf(Msg(key, lang), args...)
}
