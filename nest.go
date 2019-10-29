package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"nest/command"
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
			Usage:  "Get task dep files",
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
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
