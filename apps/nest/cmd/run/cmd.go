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
	Use:   "run <task> [task2 ...]",
	Short: "Execute one or more tasks defined in nest.yaml",
	Long: `Execute one or more named tasks from the config file.

By default, output goes directly to the terminal (raw mode).
Use --ui to launch a web-based UI with a step tree, live output, and controls.

Execution flow:
  1. Load nest.yaml (or the file specified by -c).
  2. Resolve the named task(s) and flatten all "use" references.
  3. Execute each step sequentially:
       • run:      Local bash command (supports multi-line YAML blocks).
       • use:      Inline all commands from another task (recursive).
       • upload:   Compress source to tar.gz, upload to cloud storage.
       • deploy:   For each target server:
                     a) Upload files via SCP (tar+extract) or cloud storage.
                     b) Run post-deploy commands via SSH.

Remote commands run in a login shell (bash -l) so PATH, nvm, pyenv, etc. are
available. Multi-line "run:" blocks preserve newlines.

Examples:
  nest run build                    # Run the "build" task
  nest run build deploy             # Run "build" then "deploy"
  nest run deploy --ui              # Run "deploy" with web UI
  nest run deploy -c nest.prod.yaml # Use a custom config file`,
	Run: run,
}

func init() {
	Cmd.Flags().BoolVar(&uiMode, "ui", false, "Launch web UI with step tree and live output (default: raw terminal output)")
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
			rd := taskRunner.CommandDetails()
			details := make([]webui.CommandDetail, len(rd))
			for i, d := range rd {
				details[i] = webui.CommandDetail{Name: d.Name, Depth: d.Depth, IsGroup: d.IsGroup}
			}
			webui.RunWithUI(v, taskRunner.CommandNames(), details, projectName, projectPath, common.ConfigFile, func(h webui.EventHandler, ctx context.Context) {
				// The handler implements the same interface as runner.CommandEventHandler
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
