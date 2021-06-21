package core

import (
	"github.com/koyeo/nest/logger"
)

func AddTaskRecord(branch, id, md5 string) (taskRecord *TaskRecord, err error) {

	taskRecord, err = FindTaskRecord(branch, id)
	if taskRecord != nil {
		return
	}

	taskRecord = NewTaskRecord(branch, id)
	taskRecord.Md5 = md5
	err = CreateTaskRecord(taskRecord)
	if err != nil {
		logger.Error("Create task record error: ", err)
		return
	}

	return
}

func RefreshTaskRecord(branch, id, md5 string) (err error) {

	taskRecord, err := FindTaskRecord(branch, id)
	if taskRecord == nil {
		return
	}

	taskRecord = NewTaskRecord(branch, id)
	taskRecord.Md5 = md5
	err = UpdateTaskRecord(taskRecord)

	if err != nil {
		logger.Error("Create task record error: ", err)
		return
	}

	return
}

func FindTaskRecord(branch, id string) (taskRecord *TaskRecord, err error) {

	taskRecord, err = GetTaskRecord(branch, id)
	if err != nil {
		logger.Error("GetTask task record error: ", err)
		return
	}

	return
}

func FindTaskRecords(branch string) (taskRecords []*TaskRecord, err error) {

	taskRecords, err = GetTaskRecords(branch)
	if err != nil {
		logger.Error("GetTask task records error: ", err)
		return
	}

	return
}

func CleanTaskRecord(branch, id string) (err error) {

	err = DeleteTaskRecord(branch, id)
	if err != nil {
		logger.Error("Delete task record error: ", err)
		return
	}

	return
}

func AddFileRecord(branch, ident, path, md5 string, modAt int64) (fileRecord *FileRecord, err error) {

	fileRecord, err = FindFileRecord(branch, ident)
	if fileRecord != nil {
		return
	}

	fileRecord = NewFileRecord(branch, ident, path, md5, modAt)
	err = CreateFileRecord(fileRecord)
	if err != nil {
		logger.Error("Create file record error: ", err)
		return
	}

	return
}

func FindFileRecord(branch, ident string) (fileRecord *FileRecord, err error) {

	fileRecord, err = GetFileRecord(branch, ident)
	if err != nil {
		logger.Error("GetTask file record error: ", err)
		return
	}

	return
}

func CleanFileRecord(branch, ident string) (err error) {

	err = DeleteFileRecord(branch, ident)
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

func AddFileTaskRecord(branch, fileIdent, taskId string) (err error) {

	fileTaskRecord, err := GetFileTaskRecord(branch, fileIdent, taskId)
	if err != nil {
		logger.Error("GetTask file task record error: ", err)
		return
	}

	if fileTaskRecord != nil {
		return
	}

	fileTaskRecord = NewFileTaskRecord(branch, fileIdent, taskId)
	err = CreateFileTaskRecord(fileTaskRecord)
	if err != nil {
		logger.Error("Create file task record error: ", err)
		return
	}

	return
}

func FindTaskFileRecords(branch, taskId string) (fileTaskRecords []*FileTaskRecord, err error) {
	fileTaskRecords, err = GetTaskFileRecords(branch, taskId)
	if err != nil {
		logger.Error("GetTask file task records error: ", err)
		return
	}
	return
}

func CleanTaskFileRecords(branch, taskId string) (err error) {
	err = DeleteTaskFileRecords(branch, taskId)
	if err != nil {
		logger.Error("Delete file task record error: ", err)
		return
	}
	return
}

func CleanFileTaskRecords(branch, fileIdent string) (err error) {
	err = DeleteFileTaskRecords(branch, fileIdent)
	if err != nil {
		logger.Error("Delete file tasks record error: ", err)
		return
	}
	return
}

func AddBinRecord(taskId, envId, branch, path, md5 string, modAt int64) (binRecord *BinRecord, err error) {

	binRecord, err = FindBinRecord(taskId, envId, branch)
	if binRecord != nil {
		return
	}

	binRecord = NewBinRecord(taskId, envId, branch, path, md5, modAt)
	err = CreateBinRecord(binRecord)
	if err != nil {
		logger.Error("Create bin record error: ", err)
		return
	}

	return
}

func FindBinRecord(taskId, envId, branch string) (binRecord *BinRecord, err error) {

	binRecord, err = GetBinRecord(taskId, envId, branch)
	if err != nil {
		logger.Error("GetTask bin record error: ", err)
		return
	}

	return
}

func CleanBinRecord(taskId, envId, branch string) (err error) {

	err = DeleteBinRecord(taskId, envId, branch)
	if err != nil {
		logger.Error("Delete bin record error: ", err)
		return
	}

	return
}

func RefreshBinRecord(taskId, envId, branch, md5 string, modAt int64) (err error) {

	var binRecord *BinRecord
	binRecord, err = FindBinRecord(taskId, envId, branch)
	if err != nil {
		return
	}

	binRecord.Md5 = md5
	binRecord.ModAt = modAt

	err = UpdateBinRecord(binRecord)
	if err != nil {
		logger.Error("Update bin record error: ", err)
		return
	}

	return
}
