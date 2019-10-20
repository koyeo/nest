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
	"nest/storage"
	"os"
	"path/filepath"
)

func BuildCommand(c *cli.Context) (err error) {

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
		err = execCommand(task.Task, build)
		if err != nil {
			return
		}

		log.Println(chalk.Green.Color("Build end"))
	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
		return
	}

	err = cleanBinDir()
	if err != nil {
		return
	}

	for _, task := range change.TaskList {
		if task.Type == enums.ChangeTypeDelete {
			continue
		}

		if !task.Modify {
			continue
		}
		err = moveBin(task.Task.Directory, task.Task.Id)
		if err != nil {
			return
		}
	}

	err = core.Commit(change)
	if err != nil {
		return
	}

	log.Println(chalk.Green.Color("Commit success"))

	notify.BuildDone(count)

	return
}

func execCommand(task *core.Task, build *core.Build) (err error) {
	for _, command := range build.Command {
		err = Exec(task.Directory, command)
		if err != nil {
			return
		}
	}
	return
}

func cleanBinDir() (err error) {

	binDir := storage.BinDir()

	if storage.Exist(binDir) {
		command := fmt.Sprintf("rm *")
		err = Exec(binDir, command)
		if err != nil {
			logger.Error("Clean bin error: ", err)
			return
		}
	}

	return
}

func moveBin(directory, taskId string) (err error) {

	binFile := filepath.Join(directory, enums.BuildBinConst)
	if !storage.Exist(binFile) {
		return
	}

	if !storage.Exist(storage.BinDir()) {
		storage.MakeDir(storage.BinDir())
	}

	command := fmt.Sprintf("mv __BIN__ %s", storage.BinFile(taskId))
	err = Exec(directory, command)
	if err != nil {
		return
	}
	return
}
