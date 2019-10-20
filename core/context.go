package core

import (
	"fmt"
	"nest/config"
	"nest/enums"
	"nest/storage"
)

type Context struct {
	Id        string
	Name      string
	Workspace string
	Env       []*Env
	envMap    map[string]*Env
	Script    []*Script
	scriptMap map[string]*Script
	Task      []*Task
	taskMap   map[string]*Task
}

func (p *Context) GetEnv(envId string) *Env {
	if v, ok := p.envMap[envId]; ok {
		return v
	}
	return nil
}

func (p *Context) GetTask(taskId string) *Task {
	if v, ok := p.taskMap[taskId]; ok {
		return v
	}
	return nil
}

func (p *Context) GetScript(scriptId string) *Script {
	if v, ok := p.scriptMap[scriptId]; ok {
		return v
	}
	return nil
}

func (p *Context) AddEnv(env *config.Env) (err error) {

	if p.envMap == nil {
		p.envMap = make(map[string]*Env)
	}

	if _, ok := p.envMap[env.Id]; ok {
		err = fmt.Errorf("env \"%s\" duplicated", env.Id)
		return
	}

	n, err := ToEnv(env)
	if err != nil {
		return
	}

	p.envMap[env.Id] = n
	p.Env = append(p.Env, n)

	return
}

func (p *Context) AddScript(script *config.Script) (err error) {

	if p.scriptMap == nil {
		p.scriptMap = make(map[string]*Script)
	}

	if _, ok := p.scriptMap[script.Id]; ok {
		err = fmt.Errorf("script \"%s\" duplicated", script.Id)
		return
	}

	n := ToScript(script)
	p.scriptMap[script.Id] = n
	p.Script = append(p.Script, n)

	return
}

func (p *Context) AddTask(task *config.Task) (err error) {

	if p.taskMap == nil {
		p.taskMap = make(map[string]*Task)
	}

	if _, ok := p.taskMap[task.Id]; ok {
		err = fmt.Errorf("task \"%s\" duplicated", task.Id)
		return
	}

	n, err := ToTask(task)
	if err != nil {
		return
	}
	p.taskMap[task.Id] = n
	p.Task = append(p.Task, n)

	return
}

type Env struct {
	Id        string
	Name      string
	Server    []*Server
	serverMap map[string]*Server
}

func ToEnv(o *config.Env) (n *Env, err error) {
	n = new(Env)
	n.Id = o.Id
	n.Name = o.Name
	for _, v := range o.Server {

		if n.serverMap == nil {
			n.serverMap = make(map[string]*Server)
		}
		if _, ok := n.serverMap[v.Id]; ok {
			err = fmt.Errorf("server \"%s\" duplicated", v.Id)
			return
		}
		s := ToServer(v)
		n.serverMap[v.Id] = s
		n.Server = append(n.Server, s)
	}

	return
}

type Server struct {
	Id       string
	Name     string
	Ip       string
	Port     uint64
	User     string
	Password string
	Identity string
}

func ToServer(o *config.Server) *Server {
	n := new(Server)
	n.Id = o.Id
	n.Name = o.Name
	n.Ip = o.Ip
	n.Port = o.Port
	n.User = o.User
	n.Password = o.Password
	n.Identity = o.Identity
	return n
}

type Script struct {
	Id      string
	Name    string
	File    string
	Command []string
}

func ToScript(o *config.Script) *Script {
	n := new(Script)
	n.Id = o.Id
	n.Name = o.Name
	n.Command = o.Command
	return n
}

type ExtendScript struct {
	Type    string
	File    string
	Command []string
}

func NewExtendScript(name string) (extendScript *ExtendScript, err error) {
	extendScript = new(ExtendScript)
	extendScript.File, extendScript.Type, err = config.ParseExtendScript(name)
	return
}

func (p *ExtendScript) Content() (content []byte, err error) {

	if !storage.Exist(p.File) {
		err = fmt.Errorf("sciprt file \"%s\" not exist", p.File)
		return
	}
	content, err = storage.Read(p.File)
	return
}

type Task struct {
	Id        string
	Name      string
	Watch     []string
	Directory string
	Run       string
	Build     []*Build
	buildMap  map[string]*Build
	Deploy    []*Deploy
	deployMap map[string]*Deploy
}

func (p *Task) AddBuild(build *config.Build) (err error) {
	if p.buildMap == nil {
		p.buildMap = make(map[string]*Build)
	}
	if _, ok := p.buildMap[build.Env]; ok {
		return
	}

	p.buildMap[build.Env], err = ToBuild(build)
	if err != nil {
		return
	}
	p.Build = append(p.Build, p.buildMap[build.Env])

	return
}

func (p *Task) GetBuild(env string) *Build {
	if v, ok := p.buildMap[env]; ok {
		return v
	}
	return nil
}

func (p *Task) AddDeploy(deploy *config.Deploy) (err error) {
	if p.deployMap == nil {
		p.deployMap = make(map[string]*Deploy)
	}
	if _, ok := p.deployMap[deploy.Env]; ok {
		return
	}

	p.deployMap[deploy.Env], err = ToDeploy(deploy)
	if err != nil {
		return
	}
	p.Deploy = append(p.Deploy, p.deployMap[deploy.Env])

	return
}

func (p *Task) GetDeploy(env string) *Deploy {
	if v, ok := p.deployMap[env]; ok {
		return v
	}
	return nil
}

func ToTask(o *config.Task) (n *Task, err error) {
	n = new(Task)
	n.Id = o.Id
	n.Name = o.Name
	n.Watch = o.Watch
	n.Directory = o.Directory
	n.Run = o.Run
	for _, v := range o.Build {
		err = n.AddBuild(v)
		if err != nil {
			return
		}
	}
	for _, v := range o.Deploy {
		err = n.AddDeploy(v)
		if err != nil {
			return
		}
	}
	return
}

type Build struct {
	Env          string
	BeforeScript []*ExtendScript
	AfterScript  []*ExtendScript
	Command      []string
}

func ToBuild(o *config.Build) (n *Build, err error) {
	n = new(Build)
	n.Env = o.Env
	for _, v := range o.Script {
		var extendScript *ExtendScript
		extendScript, err = NewExtendScript(v)
		if err != nil {
			return
		}
		if extendScript.Type == enums.ScriptTypeBefore {
			n.BeforeScript = append(n.BeforeScript, extendScript)
		} else if extendScript.Type == enums.ScriptTypeAfter {
			n.AfterScript = append(n.AfterScript, extendScript)
		}

	}
	n.Command = o.Command
	return
}

type Deploy struct {
	Env          string
	Log          *Log
	Pid          string
	BeforeScript []*ExtendScript
	AfterScript  []*ExtendScript
	Command      []string
}

func ToDeploy(o *config.Deploy) (n *Deploy, err error) {
	n = new(Deploy)
	n.Env = o.Env
	n.Log = ToLog(o.Log)
	n.Pid = o.Pid
	for _, v := range o.Script {
		var extendScript *ExtendScript
		extendScript, err = NewExtendScript(v)
		if err != nil {
			return
		}
		if extendScript.Type == enums.ScriptTypeBefore {
			n.BeforeScript = append(n.BeforeScript, extendScript)
		} else if extendScript.Type == enums.ScriptTypeAfter {
			n.AfterScript = append(n.AfterScript, extendScript)
		}

	}
	n.Command = o.Command
	return
}

type Log struct {
	Dir string
}

func ToLog(o *config.Log) *Log {
	n := new(Log)
	n.Dir = o.Dir
	return n
}

func MakeContext(config *config.Config) (ctx *Context, err error) {
	ctx = new(Context)
	ctx.Name = config.Name
	ctx.Workspace = config.Workspace
	for _, v := range config.Env {
		err = ctx.AddEnv(v)
		if err != nil {
			return
		}
	}
	for _, v := range config.Script {
		err = ctx.AddScript(v)
		if err != nil {
			return
		}
	}
	for _, v := range config.Task {
		err = ctx.AddTask(v)
		if err != nil {
			return
		}
	}
	return
}
