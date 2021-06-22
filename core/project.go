package core

import (
	"fmt"
	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/storage"
)

var Project = &project{}

type project struct {
	serverManager  *ServerManager
	watcherManager *WatcherManager
	taskManager    *TaskManager
}

func (p *project) ServerManager() *ServerManager {
	if p.serverManager == nil {
		p.serverManager = new(ServerManager)
	}
	return p.serverManager
}

func (p *project) WatcherManager() *WatcherManager {
	if p.watcherManager == nil {
		p.watcherManager = new(WatcherManager)
	}
	return p.watcherManager
}

func (p *project) TaskManager() *TaskManager {
	if p.taskManager == nil {
		p.taskManager = new(TaskManager)
	}
	return p.taskManager
}

func (p *project) LoadConfig() (err error) {
	for _, v := range config.Config.Servers {
		err = p.LoadServer(v)
		if err != nil {
			return
		}
	}
	for _, v := range config.Config.Watchers {
		err = p.LoadWatcher(v)
		if err != nil {
			return
		}
	}
	for _, v := range config.Config.Tasks {
		err = p.LoadTask(v)
		if err != nil {
			return
		}
	}
	err = p.LoadPipelines()
	if err != nil {
		return
	}
	return
}

func (p *project) LoadServer(conf *config.Server) (err error) {
	item := &Server{
		Name:         conf.Name,
		Host:         conf.Host,
		Port:         conf.Port,
		User:         conf.User,
		Password:     conf.Password,
		IdentityFile: conf.IdentityFile,
	}
	
	err = p.ServerManager().Add(item)
	if err != nil {
		return
	}
	
	if !storage.Exist(conf.IdentityFile) {
		err = fmt.Errorf("server '%s' identity_file '%s' not exist", item.Name, item.IdentityFile)
		return
	}
	return
}

func (p *project) LoadWatcher(conf *config.Watcher) (err error) {
	item := &Watcher{
		Name:    conf.Name,
		Command: conf.Command,
		Watch:   conf.Watch,
	}
	
	err = p.WatcherManager().Add(item)
	if err != nil {
		return
	}
	return
}

func (p *project) LoadTask(conf *config.Task) (err error) {
	item := &Task{
		Name:          conf.Name,
		BuildCommand:  Statement(conf.BuildCommand),
		DeploySource:  Statement(conf.DeploySource),
		DeployPath:    Statement(conf.DeployPath),
		DeployCommand: Statement(conf.DeployCommand),
	}
	
	err = p.TaskManager().Add(item)
	if err != nil {
		return
	}
	
	if conf.BuildScriptFile != "" {
		if !storage.Exist(conf.BuildScriptFile) {
			err = fmt.Errorf("task '%s' build_script_file '%s' not exist", item.Name, conf.BuildScriptFile)
			return
		}
		item.BuildScriptFile = conf.BuildScriptFile
	}
	
	if conf.DeployScriptFile != "" {
		if !storage.Exist(conf.DeployScriptFile) {
			err = fmt.Errorf("task '%s' deploy_script_file '%s' not exist", item.Name, conf.BuildScriptFile)
			return
		}
		item.DeployScript, err = storage.Read(conf.DeployScriptFile)
		if err != nil {
			err = fmt.Errorf("read task '%s' deploy_script_file error: %s", item.Name, err)
			return
		}
	}
	
	if conf.DeploySupervisorFile != "" {
		if !storage.Exist(conf.DeploySupervisorFile) {
			err = fmt.Errorf("task '%s' deploy_supervisor_file '%s' not exist", item.Name, conf.DeploySupervisorFile)
			return
		}
		item.DeploySupervisorConf, err = storage.Read(conf.DeploySupervisorFile)
		if err != nil {
			err = fmt.Errorf("read task '%s' deploy_supervisor_file error: %s", item.Name, err)
			return
		}
	}
	
	for _, v := range conf.DeployServer {
		server := p.ServerManager().Get(v)
		if server == nil {
			err = fmt.Errorf("task '%s' use server '%s' not exist", item.Name, v)
			return
		}
		item.DeployServer = append(item.DeployServer, server)
	}
	
	return
}

func (p *project) LoadPipelines() (err error) {
	for _, v := range config.Config.Tasks {
		self := p.TaskManager().Get(v.Name)
		if self == nil {
			err = fmt.Errorf("cant't fetch task '%s'", v.Name)
			return
		}
		for _, vv := range v.Flow {
			if vv == "_" {
				vv = v.Name
			}
			task := p.TaskManager().Get(vv)
			if task == nil {
				err = fmt.Errorf("task '%s' pipeline use task '%s' not exist", v.Name, vv)
				return
			}
			self.Flow = append(self.Flow, task)
		}
	}
	return
}
