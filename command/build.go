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
		log.Println(chalk.Green.Color("Exec directory:"), dir)

		build := task.Task.GetBuild(env.Id)
		if build == nil {
			continue
		}

		err = execCommand(ctx.Directory, task.Task, build)
		if err != nil {
			return
		}

		branch := core.Branch(ctx.Directory)

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

	err = core.CommitBuild(change)
	if err != nil {
		return
	}

	log.Println(chalk.Green.Color("Commit build success"))

	notify.BuildDone(count)

	return
}

func execCommand(projectDir string, task *core.Task, build *core.Build) (err error) {
	for _, command := range build.Command {
		err = core.PipeExec(filepath.Join(projectDir, task.Directory), command)
		if err != nil {
			return
		}
	}
	return
}

func moveBin(projectDir, taskDir, taskId, branch, envId, dist string) (err error) {

	buildFile := filepath.Join(projectDir, taskDir, dist)
	if !storage.Exist(buildFile) {
		return
	}

	binName := fmt.Sprintf("%s_%s_%s", taskId, branch, core.BuildTimestamp())

	binDir := filepath.Join(projectDir, storage.BinDir(), taskId, envId, branch)
	if storage.Exist(binDir) {
		command := fmt.Sprintf("rm -rf %s/*", binDir)
		err = core.PipeExec("", command)
		if err != nil {
			logger.Error("Clean bin error: ", err)
			return
		}
	} else {
		storage.MakeDir(binDir)
	}

	buildBinFile := filepath.Join(binDir, dist)

	err = core.PipeExec("", fmt.Sprintf("mv %s %s", buildFile, buildBinFile))
	if err != nil {
		return
	}

	err = core.PipeExec(binDir, fmt.Sprintf("zip -q  %s.zip %s", binName, dist))
	if err != nil {
		return
	}

	err = core.PipeExec("", fmt.Sprintf("rm %s", buildBinFile))
	if err != nil {
		return
	}
	return
}
