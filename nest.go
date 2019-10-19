package main

import (
	"fmt"
	"log"
	"nest/command"
	"os"

	"github.com/urfave/cli"
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
			Name:   "status",
			Usage:  "Show project status",
			Action: command.StatusCommand,
		},
		{
			Name:   "build",
			Usage:  "build project tasks",
			Action: command.BuildCommand,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
