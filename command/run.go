package command

import (
	"github.com/urfave/cli"
	"nest/core"
	"os"
)

func RunCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	task := mustGetTask(c, 0)
	if task == nil {
		return
	}

	err = core.PipeExec(task.Directory, task.Run)
	if err != nil {
		return
	}
	return
}
