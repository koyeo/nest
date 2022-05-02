package run

import (
	"fmt"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/runner"
	"github.com/spf13/cobra"
)

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "run",
	Short: "执行任务",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: exec,
}

func exec(cmd *cobra.Command, args []string) (err error) {
	conf, err := protocol.Load("nest.yml")
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
		logger.PrintStep("执行任务", v)
		err = taskRunner.Exec()
		if err != nil {
			return
		}
	}
	return
}
