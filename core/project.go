package core

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"nest/config"
	"nest/enums"
	"nest/logger"
	"nest/storage"
	"os"
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
		os.Exit(1)
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

	ctx, err = MakeContext(conf)
	if err != nil {
		logger.Error("Prepare error: ", err)
		return
	}

	context = ctx

	return
}
