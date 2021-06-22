package main

import (
	"fmt"
	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/constant"
	"github.com/koyeo/nest/protocol"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func main() {
	app := cli.NewApp()
	app.Name = "nest"
	app.Usage = "nest"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    constant.CONFIG,
			Aliases: []string{"c"},
			Usage:   "use assigned config file",
		},
	}
	app.Action = func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	}
	app.Commands = []*cli.Command{
		{
			Name:  "init",
			Usage: "init project",
			Action: func(c *cli.Context) (err error) {
				err = initConfig(c)
				if err != nil {
					return
				}
				return
			},
		},
		{
			Name:  "list",
			Usage: "list project tasks",
			Action: func(c *cli.Context) (err error) {
				err = initConfig(c)
				if err != nil {
					return
				}
				if len(protocol.Project.ServerManager().List()) > 0 {
					fmt.Println(chalk.Green.Color("servers:"))
					for _, v := range protocol.Project.ServerManager().List() {
						fmt.Printf("  %s\n", v.Name)
					}
				}
				if len(protocol.Project.WatcherManager().List()) > 0 {
					fmt.Println(chalk.Green.Color("watcher:"))
					for _, v := range protocol.Project.WatcherManager().List() {
						fmt.Printf("  %s\n", v.Name)
					}
				}
				if len(protocol.Project.TaskManager().List()) > 0 {
					fmt.Println(chalk.Green.Color("tasks:"))
					for _, v := range protocol.Project.TaskManager().List() {
						if len(v.Flow) > 0 {
							fmt.Printf("  %s(%s)\n", v.Name, chalk.Yellow.Color("pipeline"))
						} else {
							fmt.Printf("  %s\n", v.Name)
						}
					}
				}
				return
			},
		},
		{
			Name:  "watch",
			Usage: "run watcher",
			Action: func(c *cli.Context) (err error) {
				err = initConfig(c)
				if err != nil {
					return
				}
				return
			},
		},
		{
			Name:  "run",
			Usage: "run task",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    constant.FLOW,
					Aliases: []string{"f"},
					Usage:   "run task flow",
				},
			},
			Action: func(c *cli.Context) (err error) {
				err = initConfig(c)
				if err != nil {
					return
				}
				if c.Args().Len() == 0 {
					return cli.ShowCommandHelp(c, "run")
				}
				for _, name := range c.Args().Slice() {
					task := protocol.Project.TaskManager().Get(name)
					if task == nil {
						err = fmt.Errorf("task '%s' not found", name)
						return
					}
					err = task.Run(c)
					if err != nil {
						return
					}
				}
				
				return
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig(c *cli.Context) error {
	path := c.String(constant.NEST_TOML)
	if path != "" {
		return config.Load(path)
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = config.Load(filepath.Join(pwd, constant.NEST_TOML))
	if err != nil {
		return err
	}
	
	err = protocol.Project.LoadConfig()
	if err != nil {
		return err
	}
	return nil
}
