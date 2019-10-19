package core

import (
	"fmt"
	"nest/enums"
	"nest/logger"
	"nest/utils/secret"
	"os"
	"path/filepath"
	"strings"
)

type Change struct {
	TaskList []*ChangeTask
	taskMap  map[string]*ChangeTask
}

func (p *Change) Get(taskId string) *ChangeTask {
	if v, ok := (p.taskMap)[taskId]; ok {
		return v
	}
	return nil
}

func (p *Change) Add(taskId string, task *ChangeTask) {

	if p.taskMap == nil {
		p.taskMap = make(map[string]*ChangeTask)
	}

	if _, ok := (p.taskMap)[taskId]; ok {
		return
	}

	p.taskMap[taskId] = task
	p.TaskList = append(p.TaskList, task)
}

type ChangeFile struct {
	Type  int
	Ident string
	Path  string
	Md5   string
	ModAt int64
}

func NewChangeFile(path string) *ChangeFile {
	file := new(ChangeFile)
	file.Ident = secret.Md5([]byte(path))
	file.Path = path
	return file
}

type ChangeTask struct {
	Task     *Task
	Type     int
	None     []*ChangeFile
	New      []*ChangeFile
	Update   []*ChangeFile
	Delete   []*ChangeFile
	FileList []*ChangeFile
	fileMap  map[string]*ChangeFile
	Modify   bool
}

func (p *ChangeTask) Md5() string {
	items := make([]string, 0)
	for _, v := range p.FileList {
		if v.Type == enums.ChangeTypeDelete {
			continue
		}
		items = append(items, v.Ident)
	}
	return secret.Md5([]byte(strings.Join(items, "")))
}

func (p *ChangeTask) Get(ident string) *ChangeFile {
	if v, ok := (p.fileMap)[ident]; ok {
		return v
	}
	return nil
}

func NewChangeTask(task *Task) *ChangeTask {

	changeTask := new(ChangeTask)
	changeTask.Task = task
	changeTask.fileMap = make(map[string]*ChangeFile)

	return changeTask
}

func (p *ChangeTask) Add(list *[]*ChangeFile, changeFile *ChangeFile) {

	if p.fileMap == nil {
		p.fileMap = make(map[string]*ChangeFile)
	}

	if _, ok := (p.fileMap)[changeFile.Ident]; ok {
		return
	}

	(p.fileMap)[changeFile.Ident] = changeFile
	p.FileList = append(p.FileList, changeFile)
	*list = append(*list, changeFile)
}

func FileMd5(path string) (hash string, err error) {
	hash, err = secret.FileMd5(path)
	if err != nil {
		logger.Error("Get file \"%s\" md5 error", err)
		return
	}
	return
}

func MakeChange() (change *Change, err error) {

	ctx, err := Prepare()
	if err != nil {
		return
	}

	change = new(Change)

	for _, task := range ctx.Task {

		var taskRecord *TaskRecord
		taskRecord, err = FindTaskRecord(task.Id)
		if err != nil {
			return
		}

		changeTask := NewChangeTask(task)

		if taskRecord == nil {
			taskRecord = NewTaskRecord(task.Id)
			changeTask.Type = enums.ChangeTypeNew
			change.Add(task.Id, changeTask)
		} else {
			changeTask.Type = enums.ChangeTypeUpdate
			change.Add(task.Id, changeTask)
		}

		for _, glob := range task.Watch {

			var files []string
			files, err = filepath.Glob(glob)
			if err != nil {
				logger.Error("Match file error: ", err)
				return
			}

			for _, filePath := range files {

				var fileRecord *FileRecord
				fileRecord, err = FindFileRecord(GetFileIdent(filePath))
				if err != nil {
					return
				}

				var file os.FileInfo
				file, err = os.Stat(filePath)
				if err != nil {
					logger.Error(fmt.Sprintf("Get file \"%s\" status error: ", filePath), err)
					return
				}
				modAt := file.ModTime().Unix()

				changeFile := NewChangeFile(filePath)
				changeFile.ModAt = modAt

				if fileRecord == nil {
					changeFile.Type = enums.ChangeTypeNew
					changeFile.Md5, err = FileMd5(filePath)
					if err != nil {
						return
					}
					changeTask.Add(&changeTask.New, changeFile)
					continue
				}

				changeFile.Md5, err = FileMd5(filePath)
				if err != nil {
					return
				}

				if modAt <= fileRecord.ModAt || changeFile.Md5 == fileRecord.Md5 {
					changeFile.Type = enums.ChangeTypeNone
					changeTask.Add(&changeTask.None, changeFile)
					continue
				}

				changeFile.Type = enums.ChangeTypeUpdate
				changeTask.Add(&changeTask.Update, changeFile)
			}
		}

		var taskFileRecords []*FileTaskRecord
		taskFileRecords, err = FindTaskFileRecords(task.Id)
		if err != nil {
			return
		}

		for _, v := range taskFileRecords {
			changeFile := changeTask.Get(v.FileIdent)
			if changeFile != nil {
				continue
			}
			var fileRecord *FileRecord
			fileRecord, err = FindFileRecord(v.FileIdent)
			if err != nil {
				return
			}
			if fileRecord == nil {
				err = fmt.Errorf("file record \"%s\" not exist", v.FileIdent)
				return
			}
			changeFile = NewChangeFile(fileRecord.Path)
			changeFile.Type = enums.ChangeTypeDelete
			changeTask.Add(&changeTask.Delete, changeFile)
		}
	}

	taskRecords, err := FindTaskRecords()
	if err != nil {
		return
	}

	for _, v := range taskRecords {
		changeTask := change.Get(v.Id)
		if changeTask != nil {
			if changeTask.Md5() != v.Md5 {
				changeTask.Modify = true
			}
			continue
		}
		changeTask = NewChangeTask(&Task{
			Id: v.Id,
		})
		changeTask.Type = enums.ChangeTypeDelete
		change.Add(v.Id, changeTask)
	}

	return
}
