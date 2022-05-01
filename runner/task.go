package runner

import (
	"fmt"
	"github.com/gozelle/_exec"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
)

func NewTaskRunner(conf *protocol.Config, task *protocol.Task, key string) *TaskRunner {
	return &TaskRunner{conf: conf, task: task, key: key}
}

type TaskRunner struct {
	key     string
	conf    *protocol.Config
	task    *protocol.Task
	parents map[string]bool
}

func (p TaskRunner) prepareEnviron() []string {
	environ := make([]string, 0)
	envs := map[string]string{}
	for k, v := range p.conf.Envs {
		envs[k] = v
	}
	for k, v := range p.task.Envs {
		envs[k] = v
	}
	for k, v := range envs {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}
	return environ
}

func (p TaskRunner) Exec() error {
	for _, step := range p.task.Steps {
		if step.Use != "" {
			if err := p.use(step.Use); err != nil {
				return err
			}
		} else if step.Deploy != nil {
			if err := p.deploy(step.Deploy); err != nil {
				return err
			}
		} else {
			if err := p.execute(step); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p TaskRunner) use(key string) error {

	// check  circle dependency
	if p.parents != nil {
		if _, ok := p.parents[key]; ok {
			return fmt.Errorf("task: %s depend task: %s circlely", p.key, key)
		}
	}

	task, ok := p.conf.Tasks[key]
	if !ok {
		return fmt.Errorf("use task: %s not found", key)
	}
	taskRunner := NewTaskRunner(p.conf, task, key)

	// store parent task key to avoid circle dependency
	taskRunner.parents = map[string]bool{
		p.key: true,
	}
	return taskRunner.Exec()
}

func (p TaskRunner) deploy(deploy *protocol.Deploy) (err error) {
	fmt.Println("略过部署")
	return
}

func (p TaskRunner) execute(step *protocol.Step) (err error) {
	logger.Print(fmt.Sprintf("执行命令: %s\n", step.Run))
	runner := _exec.NewRunner()
	runner.AddCommand(step.Run)
	runner.SetEnviron(p.prepareEnviron())
	if p.task.Workspace != "" {
		runner.SetDir(p.task.Workspace)
	}
	err = runner.PipeOutput()
	if err != nil {
		err = fmt.Errorf("runner pipe exec error: %s", err)
		return
	}
	return
}
