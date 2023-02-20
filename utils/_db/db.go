package _db

import (
	"gitlab.forceup.in/qingyun/utils/initial"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func NewGormDB(dsn string) (*gorm.DB, func(), error) {
	
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             5 * time.Second, // Slow SQL threshold
			LogLevel:                  logger.Error,    // Log level
			IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,            // Disable color
		},
	)
	
	db, err := initial.NewGormDB(dsn, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, nil, err
	}
	
	_db, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	
	_db.SetMaxOpenConns(200)
	_db.SetConnMaxIdleTime(60 * time.Second)
	
	return db, func() {
		d, _ := db.DB()
		if d != nil {
			_ = d.Close()
		}
	}, nil
}
