package injector

import (
	"encoding/json"
	"fmt"
	"github.com/gozelle/fs"
	"github.com/koyeo/nest/hub/internal/config"
	"github.com/koyeo/nest/utils/_config"
)

func LoadConfig(file string) (conf *config.Config, err error) {
	
	path, err := fs.Lookup(file)
	if err != nil {
		return
	}
	
	fmt.Printf("load config file: %s\n", path)
	
	conf = new(config.Config)
	err = _config.UnmarshalConfigFile(path, conf)
	if err != nil {
		return
	}
	
	fmt.Println("配置文件:")
	d, _ := json.MarshalIndent(conf, "", "\t")
	fmt.Println(string(d))
	
	return
}
