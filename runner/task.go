package runner

import (
	"fmt"
	"github.com/gozelle/_color"
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

func (p TaskRunner) Exec() (err error) {
	for _, step := range p.task.Steps {
		if step.Use != "" {
			if err = p.use(step.Use); err != nil {
				return
			}
		} else if step.Deploy != nil {
			if err = p.deploy(step.Deploy); err != nil {
				return
			}
		} else {
			if err = p.execute(step); err != nil {
				return
			}
		}
	}
	return
}

func (p TaskRunner) use(key string) (err error) {

	defer func() {
		if err == nil {
			p.printExecUseEnd()
		}
	}()

	// check  circle dependency
	if p.parents != nil {
		if _, ok := p.parents[key]; ok {
			err = fmt.Errorf("task: %s depend task: %s circlely", p.key, key)
			return
		}
	}

	task, ok := p.conf.Tasks[key]
	if !ok {
		err = fmt.Errorf("use task: '%s' not found", key)
		return
	}
	taskRunner := NewTaskRunner(p.conf, task, key)

	// store parent task key to avoid circle dependency
	taskRunner.parents = map[string]bool{
		p.key: true,
	}
	p.printExecUseStart(key, task.Comment)
	err = taskRunner.Exec()
	if err != nil {
		return
	}
	return
}

func (p *TaskRunner) deploy(deploy *protocol.Deploy) (err error) {
	servers := map[string]*protocol.Server{}
	runners := make([]*ServerRunner, 0)
	defer func() {
		for _, v := range runners {
			v.Close()
		}
	}()
	for _, v := range deploy.Servers {
		if v.Use != "" {
			server, ok := p.conf.Servers[v.Use]
			if !ok {
				err = fmt.Errorf("deploy use server: '%s' not exists", v.Use)
				return
			}
			servers[server.Host] = server
		} else {
			servers[v.Host] = v
		}
	}
	for key, server := range servers {
		if server.Host == "" {
			err = fmt.Errorf("deploy server host is empty")
			return
		}
		serverRunner := NewServerRunner(p.conf, p, server, key)
		runners = append(runners, serverRunner)
		for _, mapper := range deploy.Mappers {
			err = serverRunner.Upload(mapper.Source, mapper.Target)
			if err != nil {
				return
			}
		}
	}
	for key, server := range servers {
		serverRunner := NewServerRunner(p.conf, p, server, key)
		runners = append(runners, serverRunner)
		for _, execute := range deploy.Executes {
			if execute.Run != "" {
				p.printServerExec(server, execute.Run)
				err = serverRunner.PipeExec(execute.Run)
				if err != nil {
					err = fmt.Errorf("server execute error: %s", err)
					return
				}
			}
		}
	}

	return
}

func (p TaskRunner) execute(step *protocol.Step) (err error) {
	if step.Run == "" {
		return
	}
	p.printExec(step.Run)
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

func (p TaskRunner) PrintStart() {
	logger.Step(p.key, p.task.Comment, "🕘", "start")
}

func (p TaskRunner) PrintSuccess() {
	logger.Step(p.key, p.task.Comment, "🎉", _color.New(_color.FgHiGreen).Sprint("success"))
}

func (p TaskRunner) PrintFailed() {
	logger.Step(p.key, p.task.Comment, "❌️", _color.New(_color.FgHiRed).Sprint("failed"))
}

func (p TaskRunner) printExecUseStart(key, comment string) {
	if comment != "" {
		key = comment
	}
	logger.Step(p.key, p.task.Comment, "👉", key)
}

func (p TaskRunner) printExecUseEnd() {
	logger.Step(p.key, p.task.Comment, "👈")
}

func (p TaskRunner) printServerExec(server *protocol.Server, command string) {
	logger.Step(
		p.key,
		p.task.Comment,
		"🏃",
		_color.New(_color.FgCyan).Sprintf("[%s]", server.Name()),
		_color.New(_color.FgWhite).Sprintf("%s", command),
	)
}

func (p TaskRunner) printExec(command string) {
	logger.Step(
		p.key,
		p.task.Comment,
		"🏃",
		_color.New(_color.FgWhite).Sprintf("%s", command),
	)
}
