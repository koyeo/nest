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
	//d, _ := json.MarshalIndent(Config, "", "\t")
	//fmt.Println(string(d))
	return
}

type config struct {
	Servers  []*Server  `toml:"servers,omitempty"`
	Watchers []*Watcher `toml:"watchers,omitempty"`
	Tasks    []*Task    `toml:"tasks,omitempty"`
}

type Server struct {
	Name         string `toml:"name,omitempty"`
	Host         string `toml:"host,omitempty"`
	Port         uint64 `toml:"port,omitempty"`
	User         string `toml:"user,omitempty"`
	Password     string `toml:"password,omitempty"`
	IdentityFile string `toml:"identity_file,omitempty"`
}

type Watcher struct {
	Name    string   `toml:"name,omitempty"`
	Command string   `toml:"command,omitempty"`
	Watch   []string `toml:"watch,omitempty"`
}

type Task struct {
	Name                 string   `toml:"name,omitempty"`
	BuildCommand         string   `toml:"build_command,omitempty"`
	BuildScriptFile      string   `toml:"build_script_file,omitempty"`
	DeployPath           string   `toml:"deploy_path,omitempty"`
	DeploySource         string   `toml:"deploy_source,omitempty"`
	DeployCommand        string   `toml:"deploy_command,omitempty"`
	DeployScriptFile     string   `toml:"deploy_script_file,omitempty"`
	DeploySupervisorFile string   `toml:"deploy_supervisor_file,omitempty"`
	DeployServer         []string `toml:"deploy_server,omitempty"`
	WorkFlow             []string `toml:"workflow,omitempty"`
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
