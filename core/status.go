package core

import (
	"fmt"
	"nest/logger"
	"nest/storage"
	"nest/utils/secret"
	"os"
	"path/filepath"
	"strings"
)

func Status() {

	ctx, err := Prepare()
	if err != nil {
		return
	}

	//var err error
	for _, task := range ctx.Task {

		taskRecord, err := AddTaskRecord(task.Id)
		if err != nil {
			return
		}

		taskFileMd5 := new(TaskFileMd5)

		for _, glob := range task.Watch {

			files, err := filepath.Glob(glob)
			if err != nil {
				logger.Error("Match file error: ", err)
				return
			}
			for _, filePath := range files {

				file, err := os.Stat(filePath)
				if err != nil {
					logger.Error(fmt.Sprintf("Get file \"%s\" status error: ", filePath), err)
					return
				}
				modAt := file.ModTime().Unix()

				fileRecord, err := AddFileRecord(filePath, GetFileIdent(filePath), modAt)
				if err != nil {
					return
				}

				if modAt > fileRecord.ModAt {
					err = RefreshFileRecord(filePath, fileRecord)
					if err != nil {
						return
					}
				}

				taskFileMd5.AddFile(fileRecord.Ident, fileRecord.Md5)
				//err = AddFileTaskRecord(fileRecord.Ident, taskRecord.Id)
				//if err != nil {
				//	return
				//}
			}
		}
		if taskRecord.Md5 != taskFileMd5.Md5() {
			fmt.Println("need build: ", task.Name)
		}
	}

}

type TaskFileMd5 struct {
	List []string
	Map  map[string]bool
}

func (p *TaskFileMd5) AddFile(ident, md5 string) {
	if p.Map == nil {
		p.Map = make(map[string]bool)
	}
	if _, ok := p.Map[ident]; ok {
		return
	}
	p.Map[ident] = true
	p.List = append(p.List, md5)
}

func (p *TaskFileMd5) Md5() string {
	return secret.Md5([]byte(strings.Join(p.List, "")))
}

func AddTaskRecord(id string) (taskRecord *TaskRecord, err error) {

	taskRecord, err = GetTaskRecord(id)
	if err != nil {
		logger.Error("Get task record error: ", err)
		return
	}

	if taskRecord != nil {
		return
	}

	taskRecord = NewTaskRecord(id)
	err = CreateTaskRecord(taskRecord)
	if err != nil {
		logger.Error("Create task record error: ", err)
		return
	}

	return
}

func AddFileRecord(path, ident string, modAt int64) (fileRecord *FileRecord, err error) {

	fileRecord, err = GetFileRecord(ident)
	if err != nil {
		logger.Error("Get file record error: ", err)
		return
	}

	if fileRecord != nil {
		return
	}

	data, err := storage.Read(path)
	if err != nil {
		logger.Error("Get file content error: ", err)
		return
	}

	fileRecord = NewFileRecord(ident, secret.Md5(data), modAt)
	err = CreateFileRecord(fileRecord)
	if err != nil {
		logger.Error("Create file record error: ", err)
		return
	}

	return
}

func RefreshFileRecord(path string, fileRecord *FileRecord) (err error) {

	data, err := storage.Read(path)
	if err != nil {
		logger.Error("Get file content error: ", err)
		return
	}

	fileRecord.Md5 = secret.Md5(data)
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
