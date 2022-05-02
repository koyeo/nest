package run

import (
	"fmt"
	"github.com/koyeo/nest/common"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/runner"
	"github.com/spf13/cobra"
	"os"
)

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "执行任务",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}()
	conf, err := protocol.Load(common.DefaultConfigFile)
	if err != nil {
		return
	}
	if len(args) == 0 {
		err = fmt.Errorf("miss task name, at least pass 1")
		return
	}

	for _, v := range args {
		task, ok := conf.Tasks[v]
		if !ok {
			err = fmt.Errorf("task: %s not found", v)
			return
		}
		taskRunner := runner.NewTaskRunner(conf, task, v)
		taskRunner.PrintStart()
		err = taskRunner.Exec()
		if err != nil {
			taskRunner.PrintFailed()
			return
		}
		taskRunner.PrintSuccess()
	}
	return
}
