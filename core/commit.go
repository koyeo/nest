package core

import (
	"fmt"
	"nest/enums"
	"nest/logger"
)

func Commit(change *Change) (err error) {

	session := engine.NewSession()
	defer func() {

		if err != nil {
			err = session.Rollback()
			if err != nil {
				logger.Error("Commit rollback session error: ", err)
			}
			logger.Error("Commit rollback", nil)
		}

		session.Close()
	}()

	err = session.Begin()
	if err != nil {
		logger.Error("Commit start session error: ", err)
		return
	}

	for _, task := range change.TaskList {

		switch task.Type {
		case enums.ChangeTypeNew:
			_, err = AddTaskRecord(task.Task.Id, task.Md5())
			if err != nil {
				return
			}
		case enums.ChangeTypeUpdate:
			err = RefreshTaskRecord(task.Task.Id, task.Md5())
			if err != nil {
				return
			}
		case enums.ChangeTypeDelete:
			err = CleanTaskRecord(task.Task.Id)
			if err != nil {
				return
			}
			err = CleanTaskFileRecords(task.Task.Id)
			if err != nil {
				return
			}
		}

		for _, file := range task.FileList {

			switch file.Type {
			case enums.ChangeTypeNew:
				_, err = AddFileRecord(file.Ident, file.Path, file.Md5, file.ModAt)
				if err != nil {
					return
				}
				err = AddFileTaskRecord(file.Ident, task.Task.Id)
				if err != nil {
					return
				}
			case enums.ChangeTypeUpdate:

				var fileRecord *FileRecord
				fileRecord, err = FindFileRecord(file.Ident)
				if err != nil {
					return
				}

				if fileRecord == nil {
					err = fmt.Errorf("file record is miss")
					return
				}

				fileRecord.Md5 = file.Md5
				fileRecord.ModAt = file.ModAt

				err = RefreshFileRecord(fileRecord)
				if err != nil {
					return
				}

			case enums.ChangeTypeDelete:
				err = CleanFileRecord(file.Ident)
				if err != nil {
					return
				}
				err = CleanFileTaskRecords(file.Ident)
				if err != nil {
					return
				}

			}
		}

	}

	err = session.Commit()
	if err != nil {
		logger.Error("Commit end session error: ", err)
		return
	}

	return
}
