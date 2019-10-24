package command

import (
	"github.com/urfave/cli"
	"nest/core"
	"os"
	"path/filepath"
)

func RunCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	task := mustGetTask(c, 0)
	if task == nil {
		return
	}

	err = core.PipeExec(filepath.Join(ctx.Directory, task.Directory), task.Run)
	if err != nil {
		return
	}
	return
}
