package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"nest/command"
	"nest/enums"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "nest"
	app.Usage = "nest"
	app.Action = func(c *cli.Context) error {
		fmt.Println("boom! I say!")
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "Init nest project",
			Action: command.Init,
		},
		{
			Name:   "dep",
			Usage:  "GetTask task dep files",
			Action: command.DepCommand,
		},
		{
			Name:   "status",
			Usage:  "Show project status",
			Action: command.StatusCommand,
		},
		{
			Name:   "build",
			Usage:  "build task",
			Action: command.BuildCommand,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  enums.DeployFlag,
					Usage: enums.DeployUsage,
				},
				cli.BoolFlag{
					Name:  enums.HideCommandFlag,
					Usage: enums.HideCommandUsage,
				},
			},
		},
		{
			Name:   "run",
			Usage:  "run task",
			Action: command.RunCommand,
		},
		{
			Name:   "deploy",
			Usage:  "deploy task",
			Action: command.DeployCommand,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  enums.PrintScriptFlag,
					Usage: enums.PrintScriptUsage,
				},
				cli.BoolFlag{
					Name:  enums.HideCommandFlag,
					Usage: enums.HideCommandUsage,
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
