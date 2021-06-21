package protocol

import (
	"fmt"
	"github.com/koyeo/nest/constant"
	"github.com/koyeo/nest/execer"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/storage"
	"github.com/koyeo/snowflake"
	"github.com/ttacon/chalk"
	"log"
	"os"
	"path/filepath"
)

type TaskManager struct {
	_list []*Task
	_map  map[string]*Task
}

func (p *TaskManager) Add(item *Task) error {
	
	if item.Name == "" {
		return fmt.Errorf("task with empty name")
	}
	
	if p._map == nil {
		p._map = map[string]*Task{}
	}
	
	if _, ok := p._map[item.Name]; ok {
		return fmt.Errorf("duplicated task: %s", item.Name)
	}
	
	p._map[item.Name] = item
	p._list = append(p._list, item)
	
	return nil
}

func (p *TaskManager) Get(name string) *Task {
	if p._map == nil {
		return nil
	}
	return p._map[name]
}

func (p *TaskManager) List() []*Task {
	return p._list
}

type Task struct {
	Name                 string
	Target               string
	BuildCommand         Statement
	BuildScriptFile      string
	DeployPath           string
	DeployCommand        Statement
	DeployScript         []byte
	DeploySupervisorConf []byte
	DeployServer         []*Server
	Pipeline             []*Task
}

func (p *Task) vars() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *Task) Run() (err error) {
	
	log.Println(chalk.Green.Color("[Run task]"), p.Name)
	
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	
	vars := map[string]string{}
	vars[constant.SELF] = p.Name
	vars[constant.WORKSPACE] = filepath.Join(pwd, constant.NEST_WORKSPACE, p.Name)
	if p.Target != "" {
		vars[constant.TARGET] = p.Target
	}
	vars[constant.TARGET] = filepath.Join(vars[constant.WORKSPACE], vars[constant.TARGET])
	
	defer func() {
		if err != nil {
			err = nil
			os.Exit(1)
		}
	}()
	
	if p.BuildCommand != "" {
		err = p.execBuildCommand(pwd, vars)
		if err != nil {
			return
		}
	}
	
	if p.BuildScriptFile != "" {
		err = p.execBuildScriptFile(pwd, vars)
		if err != nil {
			return
		}
	}
	
	logger.Successf("[Exec done]")
	
	return
}

func (p *Task) execBuildCommand(pwd string, vars map[string]string) (err error) {
	command, err := p.BuildCommand.Render(vars)
	if err != nil {
		return
	}
	workspace := vars[constant.WORKSPACE]
	if !storage.Exist(workspace) {
		storage.MakeDir(workspace)
	}
	err = execer.RunCommand("", pwd, command)
	if err != nil {
		return
	}
	return
}

func (p *Task) execBuildScriptFile(pwd string, vars map[string]string) (err error) {
	
	bs, err := storage.Read(p.BuildScriptFile)
	if err != nil {
		err = fmt.Errorf("task '%s' read build_script_file='%s' error: %s", p.Name, p.BuildScriptFile, err)
		return
	}
	
	content, err := Statement(bs).Render(vars)
	if err != nil {
		return
	}
	
	gen, err := snowflake.DefaultGenerator(1)
	if err != nil {
		err = fmt.Errorf("new snowflake generate error: %s", err)
		return
	}
	
	id, err := gen.Generate()
	if err != nil {
		err = fmt.Errorf("snowflake generate id error: %s", err)
		return
	}
	
	tmpName := fmt.Sprintf("%s-%s", filepath.Base(p.BuildScriptFile), id.String())
	tmpScript := filepath.Join(vars[constant.WORKSPACE], tmpName)
	err = storage.Write(tmpScript, []byte(content))
	if err != nil {
		err = fmt.Errorf("write tmp script error: %s", err)
		return
	}
	
	defer func() {
		_ = storage.Remove(tmpScript)
	}()
	
	log.Println(chalk.Green.Color("[Exec script]"), p.BuildScriptFile)
	
	err = execer.RunScript("", pwd, tmpScript)
	if err != nil {
		return
	}
	
	return
}
