package core

import (
	"nest/logger"
)

func AddTaskRecord(id, md5 string) (taskRecord *TaskRecord, err error) {

	taskRecord, err = FindTaskRecord(id)
	if taskRecord != nil {
		return
	}

	taskRecord = NewTaskRecord(id)
	taskRecord.Md5 = md5
	err = CreateTaskRecord(taskRecord)
	if err != nil {
		logger.Error("Create task record error: ", err)
		return
	}

	return
}

func RefreshTaskRecord(id, md5 string) (err error) {

	taskRecord, err := FindTaskRecord(id)
	if taskRecord == nil {
		return
	}

	taskRecord = NewTaskRecord(id)
	taskRecord.Md5 = md5
	err = UpdateTaskRecord(taskRecord)

	if err != nil {
		logger.Error("Create task record error: ", err)
		return
	}

	return
}

func FindTaskRecord(id string) (taskRecord *TaskRecord, err error) {

	taskRecord, err = GetTaskRecord(id)
	if err != nil {
		logger.Error("Get task record error: ", err)
		return
	}

	return
}

func FindTaskRecords() (taskRecords []*TaskRecord, err error) {

	taskRecords, err = GetTaskRecords()
	if err != nil {
		logger.Error("Get task records error: ", err)
		return
	}

	return
}

func CleanTaskRecord(id string) (err error) {

	err = DeleteTaskRecord(id)
	if err != nil {
		logger.Error("Delete task record error: ", err)
		return
	}

	return
}

func AddFileRecord(ident, path, md5 string, modAt int64) (fileRecord *FileRecord, err error) {

	fileRecord, err = FindFileRecord(ident)
	if fileRecord != nil {
		return
	}

	fileRecord = NewFileRecord(ident, path, md5, modAt)
	err = CreateFileRecord(fileRecord)
	if err != nil {
		logger.Error("Create file record error: ", err)
		return
	}

	return
}

func FindFileRecord(ident string) (fileRecord *FileRecord, err error) {

	fileRecord, err = GetFileRecord(ident)
	if err != nil {
		logger.Error("Get file record error: ", err)
		return
	}

	return
}

func CleanFileRecord(ident string) (err error) {

	err = DeleteFileRecord(ident)
	if err != nil {
		logger.Error("Delete file record error: ", err)
		return
	}

	return
}

func RefreshFileRecord(fileRecord *FileRecord) (err error) {

	err = UpdateFileRecord(fileRecord)
	if err != nil {
		logger.Error("Update file record error: ", err)
		return
	}

	return
}

func AddFileTaskRecord(fileIdent, taskId string) (err error) {

	fileTaskRecord, err := GetFileTaskRecord(fileIdent, taskId)
	if err != nil {
		logger.Error("Get file task record error: ", err)
		return
	}

	if fileTaskRecord != nil {
		return
	}

	fileTaskRecord = NewFileTaskRecord(fileIdent, taskId)
	err = CreateFileTaskRecord(fileTaskRecord)
	if err != nil {
		logger.Error("Create file task record error: ", err)
		return
	}

	return
}

func FindTaskFileRecords(taskId string) (fileTaskRecords []*FileTaskRecord, err error) {
	fileTaskRecords, err = GetTaskFileRecords(taskId)
	if err != nil {
		logger.Error("Get file task records error: ", err)
		return
	}
	return
}

func CleanTaskFileRecords(taskId string) (err error) {
	err = DeleteTaskFileRecords(taskId)
	if err != nil {
		logger.Error("Delete file task record error: ", err)
		return
	}
	return
}

func CleanFileTaskRecords(fileIdent string) (err error) {
	err = DeleteFileTaskRecords(fileIdent)
	if err != nil {
		logger.Error("Delete file tasks record error: ", err)
		return
	}
	return
}
