package main

import (
	"fmt"
	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/constant"
	"github.com/koyeo/nest/core"
	"github.com/koyeo/nest/logger"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"strings"
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
			Usage: "初始化项目",
			Action: func(c *cli.Context) (err error) {
				err = initConfig(c)
				if err != nil {
					return
				}
				return
			},
		},
		{
			Name:  "server",
			Usage: "服务器操作",
			Action: func(c *cli.Context) (err error) {
				return cli.ShowCommandHelp(c, "server")
			},
			Subcommands: []*cli.Command{
				{
					Name:  "test",
					Usage: "test server",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "socks",
							Aliases: []string{"s"},
						},
					},
					Action: func(c *cli.Context) (err error) {
						err = initConfig(c)
						if err != nil {
							return
						}
						if c.Args().Len() != 1 {
							return cli.ShowCommandHelp(c, "server test")
						}
						
						serverName := c.Args().First()
						
						server := core.Project.ServerManager().Get(serverName)
						if server == nil {
							return fmt.Errorf("server '%s' not found", serverName)
						}
						
						var (
							sshClient    *ssh.Client
							proxyAddress = c.String("socks")
						)
						
						if proxyAddress != "" {
							logger.Successf("[use socks] %s", proxyAddress)
							sshClient, err = core.NewProxySSHClient(proxyAddress, server)
						} else {
							sshClient, err = core.NewSSHClient(server)
						}
						if err != nil {
							return
						}
						
						defer func() {
							_ = sshClient.Close()
						}()
						
						var session *ssh.Session
						session, err = sshClient.NewSession()
						if err != nil {
							return
						}
						defer func() {
							_ = session.Close()
						}()
						var bs []byte
						bs, err = session.CombinedOutput("echo 'ok'")
						if err != nil {
							return
						}
						fmt.Println(chalk.Green.Color(strings.TrimSpace(string(bs))))
						
						return
					},
				},
			},
		},
		{
			Name:  "list",
			Usage: "列出项目任务列表",
			Action: func(c *cli.Context) (err error) {
				err = initConfig(c)
				if err != nil {
					return
				}
				if len(core.Project.ServerManager().List()) > 0 {
					fmt.Println(chalk.Green.Color("servers:"))
					for _, v := range core.Project.ServerManager().List() {
						fmt.Printf("  %s\n", v.Name)
					}
				}
				if len(core.Project.WatcherManager().List()) > 0 {
					fmt.Println(chalk.Green.Color("watcher:"))
					for _, v := range core.Project.WatcherManager().List() {
						fmt.Printf("  %s\n", v.Name)
					}
				}
				if len(core.Project.TaskManager().List()) > 0 {
					fmt.Println(chalk.Green.Color("tasks:"))
					for _, v := range core.Project.TaskManager().List() {
						if len(v.Pipeline) > 0 {
							fmt.Printf("  %s (%s)\n", v.Name, chalk.Yellow.Color("workflow"))
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
			Usage: "文件变动监听器",
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
			Usage: "执行任务",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "socks",
					Aliases: []string{"s"},
					Usage:   "use socks proxy, eg: 127.0.0.1:1081",
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
					task := core.Project.TaskManager().Get(name)
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
	// TODO 指定项目
	// TODO 指定配置文件
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = config.Load(filepath.Join(pwd, constant.NEST_TOML))
	if err != nil {
		return err
	}
	
	err = core.Project.LoadConfig()
	if err != nil {
		return err
	}
	return nil
}
