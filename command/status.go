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

	count := 0

	for _, task := range change.TaskList {

		if len(task.New) > 0 || len(task.Update) > 0 || len(task.Delete) > 0 {
			count++
			fmt.Println(strings.Repeat(" ", enums.FirstLevel), chalk.Green.Color(chalk.Bold.TextStyle(task.Task.Name)+":"))
			fmt.Printf("\n")
			for _, file := range task.New {
				fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Green.Color("new   "), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Green.Color(file.Path))
			}
			for _, file := range task.Update {
				fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Cyan.Color("update"), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Cyan.Color(file.Path))
			}
			for _, file := range task.Delete {
				fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Red.Color("delete"), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Red.Color(file.Path))
			}
			fmt.Printf("\n")
		}
	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
	}

	return
}
