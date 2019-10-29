package core

import (
	_ "github.com/mattn/go-sqlite3"
	"nest/logger"
	"nest/storage"
	"path/filepath"
	"xorm.io/core"
	"xorm.io/xorm"
)

var engine *xorm.Engine

func init() {
	_, err := Prepare()
	if err != nil {
		return
	}
	dbPath := filepath.Join(storage.DataDir())
	if !storage.Exist(dbPath) {
		storage.MakeDir(dbPath)
	}
	dbFile := filepath.Join(storage.DataFile())
	engine, err = xorm.NewEngine("sqlite3", dbFile)
	if err != nil {
		logger.Error("Open sqlite error", err)
		return
	}
	engine.SetTableMapper(core.SameMapper{})
	engine.SetColumnMapper(core.SameMapper{})
	err = engine.Sync2(new(TaskRecord), new(FileRecord), new(FileTaskRecord), new(BinRecord))
	if err != nil {
		logger.Error("Sync sqlite table error", err)
		return
	}
}

//func init() {
//	ctx, err := Prepare()
//	if err != nil {
//		return
//	}
//	dbPath := filepath.Join(ctx.Directory,storage.DataDir())
//	if !storage.Exist(dbPath) {
//		storage.MakeDir(dbPath)
//	}
//	dbFile := filepath.Join(ctx.Directory,storage.DataFile())
//
//	engine, err := xorm.NewEngine("sqlite3", dbFile)
//	if err != nil {
//		logger.Error("Open sqlite error", err)
//		os.Exit(2)
//	}
//	err = engine.Ping()
//	if err != nil {
//		logger.Error("Ping sqlite error", err)
//		os.Exit(2)
//	}
//	engine.SetTableMapper(core.SameMapper{})
//	engine.SetColumnMapper(core.SameMapper{})
//	err = engine.Sync2(new(TaskRecord), new(FileRecord), new(FileTaskRecord), new(BinRecord))
//	if err != nil {
//		logger.Error("Sync sqlite table error", err)
//		os.Exit(2)
//	}
//}
