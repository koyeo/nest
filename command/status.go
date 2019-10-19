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

	//logger.DebugPrint(change)

	for _, v1 := range change.TaskList {
		fmt.Println(strings.Repeat(" ", enums.FirstLevel), chalk.Green.Color(chalk.Bold.TextStyle(v1.Task.Name)+":"))
		fmt.Printf("\n")
		for _, v2 := range v1.New {
			fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Green.Color("new   "), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Green.Color(v2.Path))
		}
		for _, v2 := range v1.Update {
			fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Cyan.Color("update"), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Cyan.Color(v2.Path))
		}
		for _, v2 := range v1.Delete {
			fmt.Println(strings.Repeat(" ", enums.SecondLevel), chalk.Red.Color("delete"), strings.Repeat(" ", enums.StatusMarginLeft), chalk.Red.Color(v2.Path))
		}
		fmt.Printf("\n")
	}
	return
}
