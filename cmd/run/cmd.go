package run

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/koyeo/nest/common"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/runner"
	"github.com/koyeo/nest/webui"
	"github.com/spf13/cobra"
)

var rawMode bool

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run tasks / 执行任务",
	Run:   run,
}

func init() {
	Cmd.Flags().BoolVar(&rawMode, "raw", false, "Use raw console output (no GUI)")
}

func run(cmd *cobra.Command, args []string) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}()
	conf, err := protocol.Load(common.ConfigFile)
	if err != nil {
		return
	}
	if len(args) == 0 {
		err = fmt.Errorf("miss task name, at least pass 1")
		return
	}

	// Project info for webui
	projectPath, _ := os.Getwd()
	projectName := filepath.Base(projectPath)

	for _, v := range args {
		task, ok := conf.Tasks[v]
		if !ok {
			err = fmt.Errorf("task: %s not found", v)
			return
		}

		taskRunner := runner.NewTaskRunner(conf, task, v)

		if rawMode {
			// Raw mode: original console output
			taskRunner.PrintStart()
			err = taskRunner.Exec()
			if err != nil {
				taskRunner.PrintFailed()
				return
			}
			taskRunner.PrintSuccess()
		} else {
			// WebUI mode: launch webview/browser with real-time visualization
			rd := taskRunner.StepDetails()
			details := make([]webui.StepDetail, len(rd))
			for i, d := range rd {
				details[i] = webui.StepDetail{Name: d.Name, Depth: d.Depth, IsGroup: d.IsGroup}
			}
			webui.RunWithUI(v, taskRunner.StepNames(), details, projectName, projectPath, func(h webui.EventHandler) {
				// The handler implements the same interface as runner.StepEventHandler
				taskRunner.SetEventHandler(h)
				taskErr := taskRunner.Exec()
				h.OnTaskDone(taskErr)
				if taskErr != nil {
					err = taskErr
				}
			})
		}
	}
}
