package core

import (
	"nest/utils/secret"
)

func NewTaskRecord(id string) *TaskRecord {
	taskRecord := new(TaskRecord)
	taskRecord.Id = id
	return taskRecord
}

type TaskRecord struct {
	Id      string `xorm:"varchar(255) pk"`
	Md5     string `xorm:"varchar(32) not null"`
	BuildAt int64
}

func NewFileRecord(ident, md5 string, modAt int64) *FileRecord {
	fileRecord := new(FileRecord)
	fileRecord.Ident = ident
	fileRecord.Md5 = md5
	fileRecord.ModAt = modAt
	return fileRecord
}

func GetFileIdent(path string) string {
	return secret.Md5([]byte(path))
}

type FileRecord struct {
	Ident   string `xorm:"varchar(32) pk"`
	Md5     string `xorm:"varchar(32) not null"`
	ModAt   int64
	BuildAt int64
}

func NewFileTaskRecord(fileIdent, taskId string) *FileTaskRecord {
	fileTaskRecord := new(FileTaskRecord)
	fileTaskRecord.FileIdent = fileIdent
	fileTaskRecord.TaskId = taskId
	return fileTaskRecord
}

type FileTaskRecord struct {
	FileIdent string
	TaskId    string
}

func CreateTaskRecord(taskRecord *TaskRecord) (err error) {
	_, err = engine.Insert(taskRecord)
	return
}

func UpdateTaskRecord(taskRecord *TaskRecord) (err error) {
	_, err = engine.Where("id=?", taskRecord.Id).Update(taskRecord)
	return
}

func GetTaskRecord(id string) (taskRecord *TaskRecord, err error) {
	taskRecord = new(TaskRecord)
	has, err := engine.Where("id=?", id).Get(taskRecord)
	if err != nil {
		return
	}
	if !has {
		taskRecord = nil
	}
	return
}

func CreateFileRecord(fileRecord *FileRecord) (err error) {
	_, err = engine.Insert(fileRecord)
	return
}

func UpdateFileRecord(fileRecord *FileRecord) (err error) {
	_, err = engine.Where("ident=?", fileRecord.Ident).Update(fileRecord)
	return
}

func GetFileRecord(ident string) (fileRecord *FileRecord, err error) {
	fileRecord = new(FileRecord)
	has, err := engine.Where("ident=?", ident).Get(fileRecord)
	if err != nil {
		return
	}
	if !has {
		fileRecord = nil
	}
	return
}

func CreateFileTaskRecord(fileTaskRecord *FileTaskRecord) (err error) {
	_, err = engine.Insert(fileTaskRecord)
	return
}

func GetFileTaskRecord(fileIdent, taskId string) (fileTaskRecord *FileTaskRecord, err error) {
	fileTaskRecord = new(FileTaskRecord)
	has, err := engine.Where("fileIdent=? and taskId=?", fileIdent, taskId).Get(fileTaskRecord)
	if err != nil {
		return
	}
	if !has {
		fileTaskRecord = nil
	}
	return
}
