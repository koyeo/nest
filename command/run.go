package command

import (
	"fmt"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"log"
	"nest/core"
	"nest/logger"
	"nest/storage"
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

	env := mustGetEnv(c, 1)
	if env == nil {
		return
	}

	run := task.GetRun(env.Id)
	if run == nil {
		err = fmt.Errorf("env \"%s\" not found in task \"%s\" ", env.Id, task.Id)
		logger.Error("Run error: ", err)
		return
	}

	dir := filepath.Join(ctx.Directory, task.Directory)
	log.Println(chalk.Green.Color("Exec directory:"), storage.Abs(dir))
	err = core.PipeExec(dir, run.Start)
	if err != nil {
		return
	}
	return
}
