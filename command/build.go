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

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

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

		if task.Build.Type == enums.ChangeTypeDelete || !task.Build.Modify {
			continue
		}

		count++
		var dir string
		dir, err = filepath.Abs(task.Task.Directory)
		if err != nil {
			logger.Error("Modify get directory error: ", err)
			return
		}

		if task.Task.Name != "" {
			log.Println(chalk.Green.Color(fmt.Sprintf("Task: %s(%s)", task.Task.Id, task.Task.Name)))
		} else {
			log.Println(chalk.Green.Color(fmt.Sprintf("Task: %s", task.Task.Id)))
		}
		log.Println(chalk.Green.Color("Build start"))
		log.Println(chalk.Green.Color("PipeExec directory:"), dir)

		build := task.Task.GetBuild(env.Id)
		if build == nil {
			continue
		}

		err = execCommand(ctx.Directory, task.Task, build)
		if err != nil {
			return
		}

		branch := core.Branch(ctx.Directory)

		err = cleanBinDir(ctx.Directory, task.Task.Id, branch)
		if err != nil {
			return
		}

		err = moveBin(ctx.Directory, task.Task.Directory, task.Task.Id, branch, env.Id, build.Bin)
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

func execCommand(projectDir string, task *core.Task, build *core.Build) (err error) {
	for _, command := range build.Command {
		err = PipeExec(filepath.Join(projectDir, task.Directory), command)
		if err != nil {
			return
		}
	}
	return
}

func cleanBinDir(projectDir, taskId, branch string) (err error) {

	binDir := filepath.Join(projectDir, storage.BinDir(), taskId, branch)

	if storage.Exist(binDir) {
		command := fmt.Sprintf("rm *")
		err = PipeExec(binDir, command)
		if err != nil {
			logger.Error("Clean bin error: ", err)
			return
		}
	}

	return
}

func moveBin(projectDir, taskDir, taskId, branch, envId, bin string) (err error) {

	buildFile := filepath.Join(projectDir, taskDir, bin)
	if !storage.Exist(buildFile) {
		return
	}
	filename := fmt.Sprintf("%s_%s_%s", taskId, branch, core.BuildTimestamp())
	binDir := filepath.Join(projectDir, storage.BinDir(), taskId, envId, branch)
	if !storage.Exist(binDir) {
		storage.MakeDir(binDir)
	}

	command := fmt.Sprintf("mv %s %s", buildFile, filepath.Join(binDir, filename))
	err = PipeExec("", command)
	if err != nil {
		return
	}
	return
}
