package core

import (
	"nest/utils/secret"
)

func NewTaskRecord(branch, id string) *TaskRecord {
	taskRecord := new(TaskRecord)
	taskRecord.Branch = branch
	taskRecord.Id = id
	return taskRecord
}

type TaskRecord struct {
	Branch  string `xorm:"varchar(255)"`
	Id      string `xorm:"varchar(255)"`
	Md5     string `xorm:"varchar(32) not null"`
	BuildAt int64
}

func NewFileRecord(branch, ident, path, md5 string, modAt int64) *FileRecord {
	fileRecord := new(FileRecord)
	fileRecord.Branch = branch
	fileRecord.Ident = ident
	fileRecord.Md5 = md5
	fileRecord.Path = path
	fileRecord.ModAt = modAt
	return fileRecord
}

func GetFileIdent(path string) string {
	return secret.Md5([]byte(path))
}

type FileRecord struct {
	Branch  string `xorm:"varchar(255)"`
	Ident   string `xorm:"varchar(32)"`
	Md5     string `xorm:"varchar(32) not null"`
	Path    string `xorm:"text not null"`
	ModAt   int64
	BuildAt int64
}

func NewBinRecord(taskId, envId, branch, path, md5 string, modAt int64) *BinRecord {
	binRecord := new(BinRecord)
	binRecord.TaskId = taskId
	binRecord.EnvId = envId
	binRecord.Branch = branch
	binRecord.Md5 = md5
	binRecord.Path = path
	binRecord.ModAt = modAt
	return binRecord
}

type BinRecord struct {
	TaskId  string `xorm:"varchar(255)"`
	EnvId   string `xorm:"varchar(255)"`
	Branch  string `xorm:"varchar(255)"`
	Md5     string `xorm:"varchar(32) not null"`
	Path    string `xorm:"text not null"`
	ModAt   int64
	BuildAt int64
}

func NewFileTaskRecord(branch, fileIdent, taskId string) *FileTaskRecord {
	fileTaskRecord := new(FileTaskRecord)
	fileTaskRecord.Branch = branch
	fileTaskRecord.FileIdent = fileIdent
	fileTaskRecord.TaskId = taskId
	return fileTaskRecord
}

type FileTaskRecord struct {
	Branch    string
	FileIdent string
	TaskId    string
}

func CreateTaskRecord(taskRecord *TaskRecord) (err error) {
	_, err = engine.Insert(taskRecord)
	return
}

func UpdateTaskRecord(taskRecord *TaskRecord) (err error) {
	_, err = engine.Where("branch=? and id=?", taskRecord.Branch, taskRecord.Id).Update(taskRecord)
	return
}

func GetTaskRecord(branch, id string) (taskRecord *TaskRecord, err error) {

	taskRecord = new(TaskRecord)
	has, err := engine.Where("branch=? and id=?", branch, id).Get(taskRecord)
	if err != nil {
		return
	}
	if !has {
		taskRecord = nil
	}
	return
}

func GetTaskRecords(branch string) (taskRecord []*TaskRecord, err error) {
	taskRecord = make([]*TaskRecord, 0)
	err = engine.Where("branch=?", branch).Find(&taskRecord)
	if err != nil {
		return
	}
	return
}

func DeleteTaskRecord(branch, id string) (err error) {

	taskRecord := new(TaskRecord)
	_, err = engine.Where("branch=? and id=?", branch, id).Delete(taskRecord)

	return
}

func CreateFileRecord(fileRecord *FileRecord) (err error) {
	_, err = engine.Insert(fileRecord)
	return
}

func UpdateFileRecord(fileRecord *FileRecord) (err error) {
	_, err = engine.Where("branch=? and ident=?", fileRecord.Branch, fileRecord.Ident).Update(fileRecord)
	return
}

func GetFileRecord(branch, ident string) (fileRecord *FileRecord, err error) {
	fileRecord = new(FileRecord)
	has, err := engine.Where("branch=? and ident=?", branch, ident).Get(fileRecord)
	if err != nil {
		return
	}
	if !has {
		fileRecord = nil
	}
	return
}

func DeleteFileRecord(branch, ident string) (err error) {

	fileRecord := new(FileRecord)
	_, err = engine.Where("branch=? and ident=?", branch, ident).Delete(fileRecord)

	return
}

func CreateFileTaskRecord(fileTaskRecord *FileTaskRecord) (err error) {
	_, err = engine.Insert(fileTaskRecord)
	return
}

func GetFileTaskRecord(branch, fileIdent, taskId string) (fileTaskRecord *FileTaskRecord, err error) {
	fileTaskRecord = new(FileTaskRecord)
	has, err := engine.Where("branch=? and fileIdent=? and taskId=?", branch, fileIdent, taskId).Get(fileTaskRecord)
	if err != nil {
		return
	}
	if !has {
		fileTaskRecord = nil
	}
	return
}

func GetTaskFileRecords(branch, taskId string) (fileTaskRecords []*FileTaskRecord, err error) {

	fileTaskRecords = make([]*FileTaskRecord, 0)
	err = engine.Where("branch=? and taskId=?", branch, taskId).Find(&fileTaskRecords)
	if err != nil {
		return
	}
	return
}

func DeleteTaskFileRecords(branch, taskId string) (err error) {

	fileTaskRecord := new(FileTaskRecord)
	_, err = engine.Where("branch=? and taskId=?", branch, taskId).Delete(fileTaskRecord)

	return
}

func DeleteFileTaskRecords(branch, fileIdent string) (err error) {

	fileTaskRecord := new(FileTaskRecord)
	_, err = engine.Where("branch=? and fileIdent=?", branch, fileIdent).Delete(fileTaskRecord)

	return
}

func CreateBinRecord(binRecord *BinRecord) (err error) {
	_, err = engine.Insert(binRecord)
	return
}

func UpdateBinRecord(binRecord *BinRecord) (err error) {
	_, err = engine.Where("taskId=? and envId=? and branch=?", binRecord.TaskId, binRecord.EnvId, binRecord.Branch).Update(binRecord)
	return
}

func GetBinRecord(taskId, envId, branch string) (binRecord *BinRecord, err error) {
	binRecord = new(BinRecord)
	has, err := engine.Where("taskId=? and envId=? and branch=?", taskId, envId, branch).Get(binRecord)
	if err != nil {
		return
	}
	if !has {
		binRecord = nil
	}
	return
}

func DeleteBinRecord(taskId, envId, branch string) (err error) {

	binRecord := new(BinRecord)
	_, err = engine.Where("taskId=? and envId=? and branch=?", taskId, envId, branch).Delete(binRecord)

	return
}
