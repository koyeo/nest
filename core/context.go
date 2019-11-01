package core

import (
	"fmt"
	"github.com/urfave/cli"
	"nest/config"
	"nest/enums"
	"nest/logger"
	"nest/storage"
	"regexp"
	"strings"
)

type Context struct {
	Id             string
	Name           string
	Directory      string
	Env            []*Env
	envMap         map[string]*Env
	Script         []*Script
	scriptMap      map[string]*Script
	Task           []*Task
	taskMap        map[string]*Task
	uploadedScript map[string]bool
	cli            *cli.Context
}

func (p *Context) SetCli(cli *cli.Context) {
	if p.cli == nil {
		p.cli = cli
	}
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

func (p *Context) AddUploadedScript(scriptId string) (exist bool) {

	if p.uploadedScript == nil {
		p.uploadedScript = make(map[string]bool)
	}
	if _, ok := p.uploadedScript[scriptId]; ok {
		exist = true
		return
	}
	p.uploadedScript[scriptId] = true

	return
}

type Env struct {
	Id        string
	Name      string
	Server    []*Server
	serverMap map[string]*Server
}

func (p *Env) GetServer(serverId string) *Server {
	if v, ok := p.serverMap[serverId]; ok {
		return v
	}
	return nil
}

func (p *Env) AddServer(server *config.Server) (err error) {

	if p.serverMap == nil {
		p.serverMap = make(map[string]*Server)
	}

	if _, ok := p.serverMap[server.Id]; ok {
		err = fmt.Errorf("server \"%s\" duplicated", server.Id)
		return
	}

	n := ToServer(server)

	p.serverMap[server.Id] = n
	p.Server = append(p.Server, n)

	return
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
		err = n.AddServer(v)
		if err != nil {
			return
		}
	}

	return
}

type Server struct {
	Id        string
	Name      string
	Workspace string
	SSH       *SSH
}

type SSH struct {
	Ip       string
	Port     uint64
	User     string
	Password string
	Identity string
}

func ToSSH(o *config.SSH) *SSH {
	n := new(SSH)
	n.Ip = o.Ip
	n.Port = o.Port
	n.User = o.User
	n.Password = o.Password
	n.Identity = o.Identity
	return n
}

func ToServer(o *config.Server) *Server {
	n := new(Server)
	n.Id = o.Id
	n.Name = o.Name
	n.Workspace = o.Workspace
	if o.SSH != nil {
		n.SSH = ToSSH(o.SSH)
	}
	return n
}

type Script struct {
	Id      string
	Name    string
	File    string
	Command []string
	Vars    map[string]string
}

func ToScript(o *config.Script) *Script {
	n := new(Script)
	n.Id = o.Id
	n.Name = o.Name
	n.File = o.File
	n.Command = o.Command
	return n
}

func (p *Script) Copy() *Script {
	n := new(Script)
	n.Id = p.Id
	n.Name = p.Name
	n.File = p.File
	copy(n.Command, p.Command)
	n.Vars = make(map[string]string)
	for k, v := range p.Vars {
		n.Vars[k] = v
	}
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

type Run struct {
	Env   string
	Start string
}

func ToRun(o *config.Run) *Run {
	n := new(Run)
	n.Env = o.Env
	n.Start = o.Start
	return n
}

type Task struct {
	Id        string
	Name      string
	Watch     []string
	Directory string
	Run       []*Run
	runMap    map[string]*Run
	Build     []*Build
	buildMap  map[string]*Build
	Deploy    []*Deploy
	deployMap map[string]*Deploy
}

func (p *Task) AddRun(run *config.Run) {

	if p.runMap == nil {
		p.runMap = make(map[string]*Run)
	}
	if _, ok := p.runMap[run.Env]; ok {
		return
	}
	p.runMap[run.Env] = ToRun(run)
	p.Run = append(p.Run, p.runMap[run.Env])

	return
}

func (p *Task) GetRun(env string) *Run {
	if v, ok := p.runMap[env]; ok {
		return v
	}
	return nil
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

	for _, v := range o.Run {
		n.AddRun(v)
	}

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
	Id            string
	Force         bool
	Shell         string
	Env           string
	Dist          string
	BeforeScript  []*Script
	AfterScript   []*Script
	Command       []string
	BeforeCommand []string
	AfterCommand  []string
}

func ToBuild(o *config.Build) (n *Build, err error) {
	n = new(Build)
	n.Id = o.Id
	n.Force = o.Force
	n.Shell = o.Shell
	n.Env = o.Env
	n.Dist = o.Dist
	n.Command = o.Command
	return
}

type Bin struct {
	Source string
	Param  string
	Path   string
}

func ToBin(o *config.Bin) *Bin {
	n := new(Bin)
	n.Source = o.Source
	n.Param = o.Param
	n.Path = o.Path
	return n
}

type Deploy struct {
	Id            string
	Force         bool
	Env           string
	Bin           *Bin
	Daemon        *Daemon
	Server        []string
	BeforeScript  []*Script
	AfterScript   []*Script
	Command       []string
	BeforeCommand []string
	AfterCommand  []string
}

func ToDeploy(o *config.Deploy) (n *Deploy, err error) {
	n = new(Deploy)
	n.Id = o.Id
	n.Force = o.Force
	n.Env = o.Env
	if o.Bin != nil {
		n.Bin = ToBin(o.Bin)
	}
	if o.Daemon != nil {
		n.Daemon = ToDaemon(o.Daemon)
	}
	n.Server = o.Server
	for _, v := range o.Command {
		if strings.TrimSpace(v) != "" {
			n.Command = append(n.Command, v)
		}
	}

	return
}

type Daemon struct {
	Pid   string
	Start *DaemonStart
	Stop  *DaemonStop
	Log   *Log
}

func ToDaemon(o *config.Daemon) *Daemon {
	n := new(Daemon)
	if o.Start != nil {
		n.Start = ToDaemonStart(o.Start)
	}
	if o.Stop != nil {
		n.Stop = ToDaemonStop(o.Stop)
	}
	n.Pid = o.Pid
	if o.Log != nil {
		n.Log = ToLog(o.Log)
	}
	return n
}

type DaemonStart struct {
	Flag          string
	Command       []string
	BeforeCommand []string
	AfterCommand  []string
	BeforeScript  []*Script
	AfterScript   []*Script
}

func ToDaemonStart(o *config.DaemonStart) *DaemonStart {
	n := new(DaemonStart)
	for _, v := range o.Command {
		if strings.TrimSpace(v) != "" {
			n.Command = append(n.Command, v)
		}
	}
	n.Flag = o.Flag
	return n
}

type DaemonStop struct {
	Signal        string
	Flag          string
	Command       []string
	BeforeCommand []string
	AfterCommand  []string
	BeforeScript  []*Script
	AfterScript   []*Script
}

func ToDaemonStop(o *config.DaemonStop) *DaemonStop {
	n := new(DaemonStop)
	for _, v := range o.Command {
		if strings.TrimSpace(v) != "" {
			n.Command = append(n.Command, v)
		}
	}
	n.Flag = o.Flag
	for _, v := range o.Script {
		o.Script = append(o.Script, v)
	}
	return n
}

type Log struct {
	Path     string
	Name     string
	Level    string
	Size     uint64
	Backup   uint64
	Age      uint64
	Compress bool
	Daily    bool
}

func ToLog(o *config.Log) *Log {
	n := new(Log)
	n.Path = o.Path
	n.Name = o.File
	n.Level = o.Level
	n.Size = o.Size
	n.Backup = o.Backup
	n.Age = o.Age
	n.Compress = o.Compress
	n.Daily = o.Daily
	return n
}

func (p *Context) CheckScript(scriptId string) (script *Script, position string, err error) {

	rule := "(.+)" +
		enums.ScriptExtendIdent +
		`(` + enums.ScriptPositionBefore + `|` + enums.ScriptPositionAfter + `)(\` +
		enums.ScriptVarWrapFlagLeft +
		`.*\` + enums.ScriptVarWrapFlagRight + `)?`

	reg := regexp.MustCompile(rule)

	if !reg.MatchString(scriptId) {
		err = InvalidIncludeScriptRuleErr
		return
	}

	all := reg.FindAllStringSubmatch(scriptId, -1)
	if len(all) != 1 {
		err = InvalidIncludeScriptRuleErr
		return
	}

	ident := all[0]
	l := len(ident)
	if l != 4 {
		err = InvalidIncludeScriptRuleErr
		return
	}

	var vars map[string]string

	if l == 4 {
		vs := ident[3]
		if vs != "" {
			reg = regexp.MustCompile(`\(.*\)`)
			if !reg.MatchString(vs) {
				err = InvalidIncludeScriptRuleErr
				return
			}

			vs = strings.Trim(vs, enums.ScriptVarWrapFlagLeft+enums.ScriptVarWrapFlagRight)
			items := strings.Split(vs, enums.ScriptVarSplitFlag)

			for _, v := range items {

				v = strings.TrimSpace(v)
				if !strings.HasPrefix(v, enums.ScriptVarNameFlag) {
					err = InvalidIncludeScriptVarErr
					return
				}
				kv := strings.Split(v, enums.ScriptVarAssignFlag)
				if len(kv) != 2 {
					err = InvalidIncludeScriptVarErr
					return
				}

				if vars == nil {
					vars = make(map[string]string)
				}

				vars[kv[0]] = kv[1]
			}
		}
	}

	position = ident[2]

	id := ident[1]
	if script = p.GetScript(id); script == nil {
		err = ScriptNotExistErr
		return
	}

	script = script.Copy()
	script.Vars = vars

	return
}

func (p *Context) ParseScript(scriptId string, beforeCommand, afterCommand *[]string, beforeScript, afterScript *[]*Script) (err error) {

	script, position, err := p.CheckScript(scriptId)
	if err != nil {
		return
	}

	switch position {
	case enums.ScriptPositionBefore:
		for _, v := range script.Command {
			*beforeCommand = append(*beforeCommand, v)
		}
		*beforeScript = append(*beforeScript, script)
	case enums.ScriptPositionAfter:
		for _, v := range script.Command {
			*afterCommand = append(*afterCommand, v)
		}
		*afterScript = append(*afterScript, script)
	}

	return
}

func MakeContext(config *config.Config) (ctx *Context, err error) {

	defer func() {
		if err == InvalidIncludeScriptRuleErr ||
			err == InvalidIncludeScriptVarErr ||
			err == ScriptNotExistErr {
			err = IncludeScriptErr
		}
	}()

	ctx = new(Context)
	ctx.Name = config.Name
	ctx.Directory = config.Directory
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

	for _, v := range config.Task {

		task := ctx.GetTask(v.Id)

		for _, vv := range v.Build {
			build := task.GetBuild(vv.Env)
			for _, vvv := range vv.Script {
				err = ctx.ParseScript(vvv, &build.BeforeCommand, &build.AfterCommand, &build.BeforeScript, &build.AfterScript)
				if err != nil {
					logger.Error(fmt.Sprintf("Parse task \"%s\" build sciprt \"%s\" error: ", task.Id, vvv), err)
					if err == nil {
						err = IncludeScriptErr
					}
				}
			}
		}

		for _, vv := range v.Deploy {

			deploy := task.GetDeploy(vv.Env)
			for _, vvv := range vv.Script {
				err = ctx.ParseScript(vvv, &deploy.BeforeCommand, &deploy.AfterCommand, &deploy.BeforeScript, &deploy.AfterScript)
				if err != nil {
					logger.Error(fmt.Sprintf("Parse task \"%s\" deploy sciprt \"%s\" error: ", task.Id, vvv), err)
				}
			}
			if vv.Daemon == nil {
				continue
			}

			if vv.Daemon.Start != nil {
				for _, vvv := range vv.Daemon.Start.Script {
					err = ctx.ParseScript(vvv, &deploy.Daemon.Start.BeforeCommand, &deploy.Daemon.Start.AfterCommand, &deploy.Daemon.Start.BeforeScript, &deploy.Daemon.Start.AfterScript)
					if err != nil {
						logger.Error(fmt.Sprintf("Parse task \"%s\" deploy stop sciprt \"%s\" error: ", task.Id, vvv), err)
					}
				}
			}

			if vv.Daemon.Stop != nil {
				for _, vvv := range vv.Daemon.Stop.Script {
					err = ctx.ParseScript(vvv, &deploy.Daemon.Stop.BeforeCommand, &deploy.Daemon.Stop.AfterCommand, &deploy.Daemon.Stop.BeforeScript, &deploy.Daemon.Stop.AfterScript)
					if err != nil {
						logger.Error(fmt.Sprintf("Parse task \"%s\" deploy stop sciprt \"%s\" error: ", task.Id, vvv), err)
					}
				}
			}

		}

	}

	return
}
