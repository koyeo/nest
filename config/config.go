package config

import (
	"github.com/BurntSushi/toml"
)

var Config = &config{}

func Load(path string) (err error) {
	_, err = toml.DecodeFile(path, Config)
	if err != nil {
		return
	}
	return
}

type config struct {
	Servers  []*Server  `toml:"servers,omitempty"`
	Scripts  []*Script  `toml:"scripts,omitempty"`
	Watchers []*Watcher `toml:"watchers,omitempty"`
	Tasks    []*Task    `toml:"tasks,omitempty"`
}

type Server struct {
	Name         string `toml:"name,omitempty"`
	Port         uint64 `toml:"port,omitempty"`
	User         string `toml:"user,omitempty"`
	Password     string `toml:"password,omitempty"`
	IdentityFile string `toml:"identity_file,omitempty"`
	Host         string `toml:"host,omitempty"`
}

type Script struct {
	Name    string   `toml:"name,omitempty"`
	File    string   `toml:"file,omitempty"`
	Command []string `toml:"command,omitempty"`
}

type Watcher struct {
	Command string `toml:"command,omitempty"`
}

type Task struct {
	Name     string   `toml:"name,omitempty"`
	Build    *Build   `json:"build"`
	Deploy   *Deploy  `json:"deploy"`
	Pipeline []string `toml:"pipeline"`
}

type Build struct {
	Command string `toml:"command,omitempty"`
	Script  string `toml:"script,omitempty"`
}

type Deploy struct {
	Path    string `json:"path"`
	Server  string `toml:"server,omitempty"`
	Command string `toml:"command,omitempty"`
	Script  string `toml:"script,omitempty"`
}

//func ParseExtendScript(name string) (script, position string, err error) {
//	if strings.Contains(name, enums.ScriptExtendIdent) {
//		items := strings.Split(name, enums.ScriptExtendIdent)
//		script, position = items[0], items[1]
//		if position != enums.ScriptTypeBefore && position != enums.ScriptTypeAfter {
//			err = fmt.Errorf("invaild script position \"%s\"", position)
//			return
//		}
//	}
//	return
//}
