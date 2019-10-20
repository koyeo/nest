package command

import (
	"fmt"
	"github.com/urfave/cli"
	"nest/core"
	"nest/logger"
)

func getTask(c *cli.Context, index int) (task *core.Task) {

	args := c.Args()
	if len(args) < index+1 {
		return
	}

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	task = ctx.GetTask(args[index])
	if task == nil {
		return
	}

	return
}

func mustGetTask(c *cli.Context, index int) (task *core.Task) {

	args := c.Args()
	if len(args) < index+1 {
		logger.Error("Please input task id", nil)
		return
	}

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	task = ctx.GetTask(args[index])
	if task == nil {
		logger.Error(fmt.Sprintf("task \"%s\" not exist", args[index]), nil)
		return
	}

	return
}

func getEnv(c *cli.Context, index int) (env *core.Env) {

	args := c.Args()
	if len(args) < index+1 {
		return
	}

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	env = ctx.GetEnv(args[index])
	if env == nil {
		return
	}

	return
}

func mustGetEnv(c *cli.Context, index int) (env *core.Env) {

	args := c.Args()
	if len(args) < index+1 {
		logger.Error("Please input env id", nil)
		return
	}

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	env = ctx.GetEnv(args[index])
	if env == nil {
		logger.Error(fmt.Sprintf("env \"%s\" not exist", args[index]), nil)
		return
	}

	return
}
