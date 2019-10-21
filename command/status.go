package command

import (
	"fmt"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"nest/core"
	"nest/enums"
	"os"
	"strings"
)

func StatusCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	change, err := core.MakeChange()
	if err != nil {
		return
	}
	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	branch := core.Branch(ctx.Directory)
	if branch != "" {
		fmt.Println(chalk.Green.Color(chalk.Bold.TextStyle("Branch: " + branch)))
	}

	buildCount := 0
	deployCount := 0
	for _, task := range change.TaskList {

		build := task.Build
		if len(build.New) > 0 || len(build.Update) > 0 || len(build.Delete) > 0 {
			if buildCount == 0 {
				fmt.Println(chalk.Green.Color(chalk.Bold.TextStyle("Bin task:")))
			}
			if task.Task.Name != "" {
				fmt.Println(strings.Repeat(" ", enums.FirstLevel), chalk.Green.Color(chalk.Bold.TextStyle(fmt.Sprintf("%s (%s)", task.Task.Id, task.Task.Name))+":"))
			} else {
				fmt.Println(strings.Repeat(" ", enums.FirstLevel), chalk.Green.Color(chalk.Bold.TextStyle(task.Task.Id)+":"))
			}
			for _, file := range build.New {
				fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Green.Color("new   "), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Green.Color(file.Path))
			}
			for _, file := range build.Update {
				fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Cyan.Color("update"), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Cyan.Color(file.Path))
			}
			for _, file := range build.Delete {
				fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Red.Color("delete"), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Red.Color(file.Path))
			}
			buildCount++
		}

		for _, deploy := range task.Deploy {
			if !deploy.Modify {
				continue
			}
			if deployCount == 0 {
				fmt.Println(chalk.Green.Color(chalk.Bold.TextStyle("Deploy task:")))
				if task.Task.Name != "" {
					fmt.Println(strings.Repeat(" ", enums.FirstLevel), chalk.Green.Color(chalk.Bold.TextStyle(fmt.Sprintf("%s (%s)", task.Task.Id, task.Task.Name))+":"))
				} else {
					fmt.Println(strings.Repeat(" ", enums.FirstLevel), chalk.Green.Color(chalk.Bold.TextStyle(task.Task.Id)+":"))
				}
			}
			fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Cyan.Color(deploy.TaskId), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Cyan.Color(deploy.EnvId))
			deployCount++

		}
	}

	if buildCount == 0 && deployCount == 0 {
		fmt.Println(chalk.Green.Color("no change"))
	}

	return
}
