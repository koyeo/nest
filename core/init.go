package core

import (
	_ "github.com/mattn/go-sqlite3"
	"nest/logger"
	"xorm.io/xorm"
)

var engine *xorm.Engine

func init() {
	var err error
	engine, err = xorm.NewEngine("sqlite3", "test/.nest/nest.db")
	if err != nil {
		logger.Error("Open sqlite error", err)
		return
	}
	err = engine.Sync2(new(TaskRecord), new(FileRecord), new(FileTaskRecord))
	if err != nil {
		logger.Error("Sync sqlite table error", err)
		return
	}
}
