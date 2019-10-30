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
	ctx.SetCli(c)

	change, err := core.MakeChange()
	if err != nil {
		return
	}

	env := mustGetEnv(c, 0)
	if env == nil {
		return
	}

	count := 0
	for _, task := range ctx.Task {

		build := task.GetBuild(env.Id)
		if build == nil {
			continue
		}

		changeTask := change.GetTask(task.Id)
		if !build.Force {
			if changeTask == nil || changeTask.Build.Type == enums.ChangeTypeDelete || !changeTask.Build.Modify {
				continue
			}
		}

		count++
		var dir string
		dir, err = filepath.Abs(filepath.Join(ctx.Directory, task.Directory))
		if err != nil {
			logger.Error("Modify get directory error: ", err)
			return
		}

		if task.Name != "" {
			log.Println(chalk.Green.Color(fmt.Sprintf("Task: %s(%s)", task.Id, task.Name)))
		} else {
			log.Println(chalk.Green.Color(fmt.Sprintf("Task: %s", task.Id)))
		}
		log.Println(chalk.Green.Color("Build start"))
		log.Println(chalk.Green.Color("Exec directory:"), dir)

		err = execCommand(dir, task, build)
		if err != nil {
			return
		}

		if build.Force {
			continue
		}

		branch := core.Branch(ctx.Directory)

		err = moveBin(ctx.Directory, task.Directory, task.Id, branch, env.Id, build.Bin)
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

func execCommand(dir string, task *core.Task, build *core.Build) (err error) {
	for _, command := range build.Command {
		err = core.PipeRun(dir, command)
		if err != nil {
			return
		}
	}
	return
}

func moveBin(projectDir, taskDir, taskId, branch, envId, dist string) (err error) {

	if dist == "" {
		return
	}

	buildFile := filepath.Join(projectDir, taskDir, dist)
	if !storage.Exist(buildFile) {
		return
	}

	binName := fmt.Sprintf("%s_%s_%s", taskId, branch, core.BuildTimestamp())

	binDir := filepath.Join(storage.BinDir(), taskId, envId, branch)
	if storage.Exist(binDir) {
		command := fmt.Sprintf("rm -rf %s/*", binDir)
		err = core.PipeRun("", command)
		if err != nil {
			logger.Error("Clean bin error: ", err)
			return
		}
	} else {
		storage.MakeDir(binDir)
	}

	buildBinFile := filepath.Join(binDir, dist)

	err = core.PipeRun("", fmt.Sprintf("mv %s %s", buildFile, buildBinFile))
	if err != nil {
		return
	}

	err = core.PipeRun(binDir, fmt.Sprintf("zip -q  %s.zip %s", binName, dist))
	if err != nil {
		return
	}

	err = core.PipeRun("", fmt.Sprintf("rm %s", buildBinFile))
	if err != nil {
		return
	}
	return
}
