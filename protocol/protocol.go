package protocol

type Config struct {
	Version string             `yaml:"version"`
	Servers map[string]*Server `yaml:"servers"`
	Envs    map[string]string  `yaml:"envs"`
	Tasks   map[string]*Task   `yaml:"tasks"`
}

type Server struct {
	Alias        string `yaml:"alias"`
	Use          string `yaml:"use"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	IdentityFile string `yaml:"identity_file"`
}

type Task struct {
	Comment   string            `yaml:"comment"`
	Workspace string            `yaml:"workspace"`
	Branches  []string          `yaml:"branches"`
	Envs      map[string]string `yaml:"envs"`
	Steps     []*Step           `yaml:"steps"`
}

type Step struct {
	Comment string  `yaml:"comment"`
	Use     string  `yaml:"use"`
	Run     string  `yaml:"run"`
	Deploy  *Deploy `yaml:"deploy"`
}

type Execute struct {
	Comment string `yaml:"comment"`
	Use     string `yaml:"use"`
	Run     string `yaml:"run"`
}

type Deploy struct {
	Servers  []*Server  `yaml:"servers"`
	Mappers  []*Mapper  `yaml:"mappers"`
	Executes []*Execute `yaml:"executes"`
}

type Mapper struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}
