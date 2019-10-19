package config

type Config struct {
	Name      string    `yaml:"name,omitempty"`
	Id        string    `yaml:"id,omitempty"`
	Workspace string    `yaml:"workspace,omitempty"`
	Env       []*Env    `yaml:"env,omitempty"`
	Script    []*Script `yaml:"script,omitempty"`
	Task      []*Task   `yaml:"task,omitempty"`
}

type Env struct {
	Id     string    `yaml:"id,omitempty"`
	Name   string    `yaml:"name,omitempty"`
	Server []*Server `yaml:"server,omitempty"`
}

type Server struct {
	Id       string `yaml:"id,omitempty"`
	Name     string `yaml:"name,omitempty"`
	Ip       string `yaml:"ip,omitempty"`
	Port     uint64 `yaml:"port,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Identity string `yaml:"identity,omitempty"`
}

type Script struct {
	Id      string   `yaml:"id,omitempty"`
	Name    string   `yaml:"name,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

type Task struct {
	Id     string    `yaml:"id,omitempty"`
	Name   string    `yaml:"name,omitempty"`
	Watch  []string  `yaml:"watch,omitempty"`
	Run    string    `yaml:"run,omitempty"`
	Build  *Build    `yaml:"build,omitempty"`
	Deploy []*Deploy `yaml:"deploy,omitempty"`
}

type Build struct {
	Command []string `yaml:"command,omitempty"`
}

type Deploy struct {
	Env     string   `yaml:"env,omitempty"`
	Log     *Log     `yaml:"log,omitempty"`
	Pid     string   `yaml:"pid,omitempty"`
	Script  []string `yaml:"script,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

type Log struct {
	Dir string `yaml:"dir,omitempty"`
}
