package command

import (
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"log"
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

	dir := filepath.Join(ctx.Directory, task.Directory)
	log.Println(chalk.Green.Color("Exec directory:"), dir)
	err = core.PipeExec(dir, task.Run)
	if err != nil {
		return
	}
	return
}
