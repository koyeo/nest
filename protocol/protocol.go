package protocol

type Protocol struct {
	Version string
	Servers map[string]*Server
	Envs    map[string]string
	Tasks   map[string]*Task
}

type Server struct {
	Alias        string
	Use          string
	Host         string
	User         string
	Password     string
	IdentifyFile string
}

type Task struct {
	Comment  string
	Branches []string
	envs     map[string]string
	Steps    []interface{}
}
