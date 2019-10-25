package config

import (
	"fmt"
	"nest/enums"
	"strings"
)

type Config struct {
	Id        string    `yaml:"id,omitempty"`
	Name      string    `yaml:"name,omitempty"`
	Directory string    `yaml:"directory,omitempty"`
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
	Id   string `yaml:"id,omitempty"`
	Name string `yaml:"name,omitempty"`
	Ip   string `yaml:"ip,omitempty"`
	SSH  *SSH   `yaml:"ssh"`
}
type SSH struct {
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
	Id        string   `yaml:"id,omitempty"`
	Name      string   `yaml:"name,omitempty"`
	Watch     []string `yaml:"watch,omitempty"`
	Directory string   `yaml:"directory,omitempty"`
	Run       []*Run   `yaml:"run,omitempty"`
	Build     []*Build  `yaml:"build,omitempty"`
	Deploy    []*Deploy `yaml:"deploy,omitempty"`
}

type Run struct {
	Env   string `yaml:"env,omitempty"`
	Start string `yaml:"start,omitempty"`
}

type Build struct {
	Id      string   `yaml:"id,omitempty"`
	Env     string   `yaml:"env,omitempty"`
	Dist    string   `yaml:"dist,omitempty"`
	Script  []string `yaml:"script,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

type Deploy struct {
	Id      string   `yaml:"id,omitempty"`
	Env     string   `yaml:"env,omitempty"`
	Bin     *Bin     `yaml:"bin,omitempty"`
	Daemon  *Daemon  `yaml:"daemon,omitempty"`
	Server  []string `yaml:"server,omitempty"`
	Script  []string `yaml:"script,omitempty"`
	Command []string `yaml:"command,omitempty"`
}

type Bin struct {
	Source string `yaml:"source,omitempty"`
	Param  string `yaml:"param,omitempty"`
	Path   string `yaml:"path,omitempty"`
}

type Daemon struct {
	Start string `yaml:"start,omitempty"`
	Pid   string `yaml:"pid,omitempty"`
	Log   *Log   `yaml:"log,omitempty"`
}

type Log struct {
	Path     string `yaml:"path,omitempty"`
	File     string `yaml:"file,omitempty"`
	Level    string `yaml:"level,omitempty"`
	Size     uint64 `yaml:"size,omitempty"`
	Backup   uint64 `yaml:"backup,omitempty"`
	Age      uint64 `yaml:"age,omitempty"`
	Compress bool   `yaml:"compress,omitempty"`
	Daily    bool   `yaml:"daily,omitempty"`
}

func ParseExtendScript(name string) (script, position string, err error) {
	if strings.Contains(name, enums.ScriptExtendIdent) {
		items := strings.Split(name, enums.ScriptExtendIdent)
		script, position = items[0], items[1]
		if position != enums.ScriptTypeBefore && position != enums.ScriptTypeAfter {
			err = fmt.Errorf("invaild script position \"%s\"", position)
			return
		}
	}
	return
}
