package command

import (
	"fmt"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"log"
	"nest/core"
	"nest/enums"
	"nest/logger"
	"nest/notify"
	"os"
	"path/filepath"
)

func DeployCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	change, err := core.MakeChange()
	if err != nil {
		return
	}

	env := mustGetEnv(c, 0)
	if env == nil {
		return
	}

	count := 0
	for _, task := range change.TaskList {

		if task.Type == enums.ChangeTypeDelete {
			continue
		}

		if !task.Modify {
			continue
		}
		count++
		var dir string
		dir, err = filepath.Abs(task.Task.Directory)
		if err != nil {
			logger.Error("Modify get directory error: ", err)
			return
		}

		log.Println(chalk.Green.Color("Task:"), task.Task.Name)
		log.Println(chalk.Green.Color("Build start"))
		log.Println(chalk.Green.Color("Exec directory:"), dir)

		build := task.Task.GetBuild(env.Id)
		if build == nil {
			err = fmt.Errorf("task \"%s\"env \"%s\" not exist", task.Task.Id, env.Id)
			logger.Error("Build error: ", err)
			return
		}
		err = execBuildCommand(task.Task, build)
		if err != nil {
			return
		}

		log.Println(chalk.Green.Color("Build end"))
	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
		return
	}

	err = core.Commit(change)
	if err != nil {
		return
	}
	log.Println(chalk.Green.Color("Commit success"))

	notify.BuildDone(count)

	return
}

