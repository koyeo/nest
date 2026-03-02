package runner

import (
	"fmt"
	"strings"

	"github.com/koyeo/nest/deploy/domain"
	"github.com/koyeo/nest/i18n"
)

// HandlerPrompter adapts StepEventHandler.Prompt to domain.UserPrompter.
// Used in webui mode where stdin is not available.
type HandlerPrompter struct {
	Handler StepEventHandler
}

func (p *HandlerPrompter) AskConflictAction(files []string, lang string) (domain.ConflictAction, string, error) {
	// Build the full prompt message
	msg := fmt.Sprintf("%s\n%s",
		i18n.Msgf(i18n.MsgConflictFound, lang, strings.Join(files, ", ")),
		i18n.Msg(i18n.MsgChooseAction, lang),
	)

	choice := strings.TrimSpace(p.Handler.Prompt(msg))

	if choice == "2" {
		return domain.ActionRemove, "", nil
	}

	// Default to backup — ask for suffix
	suffix := strings.TrimSpace(p.Handler.Prompt(i18n.Msg(i18n.MsgBackupSuffix, lang)))
	if suffix == "" {
		suffix = ".bak"
	}

	return domain.ActionBackup, suffix, nil
}
