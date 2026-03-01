package infrastructure

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/koyeo/nest/deploy/domain"
	"github.com/koyeo/nest/i18n"
)

// StdinPrompter implements domain.UserPrompter using stdin/stdout.
type StdinPrompter struct{}

// NewStdinPrompter creates a new StdinPrompter.
func NewStdinPrompter() *StdinPrompter {
	return &StdinPrompter{}
}

func (p *StdinPrompter) AskConflictAction(files []string, lang string) (domain.ConflictAction, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println(i18n.Msgf(i18n.MsgConflictFound, lang, strings.Join(files, ", ")))
	fmt.Print(i18n.Msg(i18n.MsgChooseAction, lang))

	line, err := reader.ReadString('\n')
	if err != nil {
		return domain.ActionBackup, ".bak", fmt.Errorf("read input error: %s", err)
	}
	choice := strings.TrimSpace(line)

	if choice == "2" {
		return domain.ActionRemove, "", nil
	}

	// Default to backup — ask for suffix
	fmt.Print(i18n.Msg(i18n.MsgBackupSuffix, lang))
	line, err = reader.ReadString('\n')
	if err != nil {
		return domain.ActionBackup, ".bak", nil
	}
	suffix := strings.TrimSpace(line)
	if suffix == "" {
		suffix = ".bak"
	}

	return domain.ActionBackup, suffix, nil
}
