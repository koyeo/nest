package protocol

import "fmt"

const (
	Version = "1.0"
)

type Config struct {
	Version  string             `yaml:"version"`
	Servers  map[string]*Server `yaml:"servers"`
	Storages map[string]string  `yaml:"storages"` // alias → global storage config name
	Envs     map[string]string  `yaml:"envs"`
	Tasks    map[string]*Task   `yaml:"tasks"`
}

// ResolveStorage returns the global storage config name for a declared alias.
// Returns an error if the alias is not declared in the storage section.
func (c *Config) ResolveStorage(alias string) (string, error) {
	if len(c.Storages) == 0 {
		return "", fmt.Errorf("no storage declared in nest.yaml, add a 'storage' section first")
	}
	name, ok := c.Storages[alias]
	if !ok {
		return "", fmt.Errorf("storage '%s' not declared in nest.yaml", alias)
	}
	return name, nil
}

type Server struct {
	Alias        string `yaml:"alias"`
	Comment      string `yaml:"comment"`
	Use          string `yaml:"use"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	IdentityFile string `yaml:"identity_file"`
}

func (p Server) Name() string {
	serverName := p.Host
	if p.Comment != "" {
		serverName = fmt.Sprintf("%s:%s", p.Comment, serverName)
	}
	return serverName
}

type Task struct {
	Comment   string            `yaml:"comment"`
	Workspace string            `yaml:"workspace"`
	Branches  []string          `yaml:"branches"`
	Envs      map[string]string `yaml:"envs"`
	Commands  []*Command        `yaml:"commands"`
}

type Command struct {
	Comment string  `yaml:"comment"`
	Use     string  `yaml:"use"`
	Run     string  `yaml:"run"`
	Deploy  *Deploy `yaml:"deploy"`
	Upload  *Upload `yaml:"upload"`
}

// Upload defines a command that compresses and uploads artifacts to cloud storage.
type Upload struct {
	Storage string `yaml:"storage"` // reference key to global storage config
	Source  string `yaml:"source"`  // local file or directory to upload
}

type Deploy struct {
	Servers          []*Server      `yaml:"servers"`
	Files            []*FileMapping `yaml:"files"`
	Commands         []*Command     `yaml:"commands"`
	Cwd              string         `yaml:"cwd"`
	ShellInit        string         `yaml:"shell_init"`
	ConflictStrategy string         `yaml:"conflict_strategy"` // overwrite | backup | error
}

type FileMapping struct {
	Source  string `yaml:"source"`
	Target  string `yaml:"target"`
	Storage string `yaml:"storage"` // storage alias name; empty = direct SFTP transfer
}
