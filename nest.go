package main

import (
	"fmt"
	"github.com/koyeo/nest/command"
	"github.com/koyeo/nest/enums"
	"github.com/urfave/cli"
	"log"
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
			Usage:  "nest init",
			Action: command.Init,
		},
		{
			Name:   "status",
			Usage:  "nest status",
			Action: command.StatusCommand,
		},
		{
			Name:   "run",
			Usage:  "nest run task",
			Action: command.RunCommand,
		},
		{
			Name:   "build",
			Usage:  "nest build task ...",
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
			Name:   "deploy",
			Usage:  "nest deploy task ...",
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
		{
			Name:   "pipeline",
			Usage:  "exec pipeline task",
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
