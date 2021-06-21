package core

import (
	"fmt"
	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/enums"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/storage"
	"gopkg.in/yaml.v2"
	"path/filepath"
)

var context *Context

func Prepare() (ctx *Context, err error) {

	if context != nil {
		ctx = context
		return
	}

	file := filepath.Join(enums.ConfigFile)
	if !storage.Exist(file) {
		err = fmt.Errorf("%s not exist", enums.ConfigFile)
		logger.Error("Load config error: ", err)
		//os.Exit(1)
		return
	}

	conf := new(config.Config)
	data, err := storage.Read(file)
	if err != nil {
		logger.Error("Prepare read config error: ", err)
		return
	}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		logger.Error("Prepare unmarshal config error: ", err)
		return
	}

	context, err = MakeContext(conf)
	if err != nil {
		if err != IncludeScriptErr {
			logger.Error("Prepare error: ", err)
		}
		return
	}

	ctx = context

	return
}
