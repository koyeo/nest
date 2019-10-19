package core

import (
	"fmt"
	"nest/config"
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

	n := ToTask(task)
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

		if n.serverMap == nil{
			n.serverMap = make( map[string]*Server)
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
	Command []string
}

func ToScript(o *config.Script) *Script {
	n := new(Script)
	n.Id = o.Id
	n.Name = o.Name
	n.Command = o.Command
	return n
}

type Task struct {
	Id     string
	Name   string
	Watch  []string
	Run    string
	Build  *Build
	Deploy []*Deploy
}

func ToTask(o *config.Task) *Task {
	n := new(Task)
	n.Id = o.Id
	n.Name = o.Name
	n.Watch = o.Watch
	n.Run = o.Run
	n.Build = ToBuild(o.Build)
	for _, v := range o.Deploy {
		n.Deploy = append(n.Deploy, ToDeploy(v))
	}
	return n
}

type Build struct {
	Command []string
}

func ToBuild(o *config.Build) *Build {
	n := new(Build)
	n.Command = o.Command
	return n
}

type Deploy struct {
	Env     string
	Log     *Log
	Pid     string
	Script  []string
	Command []string
}

func ToDeploy(o *config.Deploy) *Deploy {
	n := new(Deploy)
	n.Env = o.Env
	n.Log = ToLog(o.Log)
	n.Pid = o.Pid
	n.Script = o.Script
	n.Command = o.Command
	return n
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
