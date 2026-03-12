package run

import (
	"context"
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

var uiMode bool

var Cmd = &cobra.Command{
	Use:   "run",
	Short: "Run tasks / 执行任务",
	Run:   run,
}

func init() {
	Cmd.Flags().BoolVar(&uiMode, "ui", false, "Use web UI for real-time visualization")
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

		if uiMode {
			// WebUI mode: launch webview/browser with real-time visualization
			rd := taskRunner.StepDetails()
			details := make([]webui.StepDetail, len(rd))
			for i, d := range rd {
				details[i] = webui.StepDetail{Name: d.Name, Depth: d.Depth, IsGroup: d.IsGroup}
			}
			webui.RunWithUI(v, taskRunner.StepNames(), details, projectName, projectPath, common.ConfigFile, func(h webui.EventHandler, ctx context.Context) {
				// The handler implements the same interface as runner.StepEventHandler
				taskRunner.SetEventHandler(h)
				taskRunner.SetContext(ctx)
				taskErr := taskRunner.Exec()
				h.OnTaskDone(taskErr)
				if taskErr != nil {
					err = taskErr
				}
			})
		} else {
			// Raw mode (default): original console output
			taskRunner.PrintStart()
			err = taskRunner.Exec()
			if err != nil {
				taskRunner.PrintFailed()
				return
			}
			taskRunner.PrintSuccess()
		}
	}
}
