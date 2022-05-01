package protocol

import (
	"fmt"
	"github.com/gozelle/_fs"
	"gopkg.in/yaml.v3"
)

func Load(path string) (config *Config, err error) {
	ok, err := _fs.Exist(path)
	if err != nil {
		return
	}
	if !ok {
		err = fmt.Errorf("path: %s not exist", path)
		return
	}

	content, err := _fs.Read(path)
	if err != nil {
		return
	}
	config = new(Config)
	err = yaml.Unmarshal(content, config)
	if err != nil {
		err = fmt.Errorf("unmarshal yml error: %s", err)
		return
	}
	return
}
