package command

import (
	"fmt"
	"github.com/urfave/cli"
	"nest/core"
	"nest/logger"
	"os"
)

func DepCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	task := mustGetTask(c, 0)
	if task == nil {
		return
	}
	
	files, err := core.TaskGlobFiles(ctx, task)
	if err != nil {
		logger.Error("Match files error: ", err)
		return
	}

	for _, v := range files {
		fmt.Println(v)
	}

	return
}
