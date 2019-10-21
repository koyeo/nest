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

	ctx, err := Prepare()
	if err != nil {
		return
	}

	branch := Branch(ctx.Directory)

	for _, task := range change.TaskList {

		build := task.Build

		switch build.Type {
		case enums.ChangeTypeNew:
			_, err = AddTaskRecord(branch, task.Task.Id, build.Md5)
			if err != nil {
				return
			}
		case enums.ChangeTypeUpdate:
			err = RefreshTaskRecord(branch, task.Task.Id, build.Md5)
			if err != nil {
				return
			}
		case enums.ChangeTypeDelete:
			err = CleanTaskRecord(branch, task.Task.Id)
			if err != nil {
				return
			}
			err = CleanTaskFileRecords(branch, task.Task.Id)
			if err != nil {
				return
			}
		}

		for _, file := range build.FileList {

			switch file.Type {
			case enums.ChangeTypeNew:
				_, err = AddFileRecord(branch, file.Ident, file.Path, file.Md5, file.ModAt)
				if err != nil {
					return
				}
				err = AddFileTaskRecord(branch, file.Ident, task.Task.Id)
				if err != nil {
					return
				}
			case enums.ChangeTypeUpdate:

				var fileRecord *FileRecord
				fileRecord, err = FindFileRecord(branch, file.Ident)
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
				err = CleanFileRecord(branch, file.Ident)
				if err != nil {
					return
				}
				err = CleanFileTaskRecords(branch, file.Ident)
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
