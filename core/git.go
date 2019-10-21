package core

import "strings"

func Branch(projectDir string) (branch string) {
	branch, _ = Exec(projectDir, "git rev-parse --abbrev-ref HEAD")
	branch = strings.TrimSpace(branch)
	return
}
