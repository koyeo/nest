package core

import (
	"fmt"
	"nest/enums"
	"nest/logger"
	"nest/storage"
	"nest/utils/secret"
	"os"
	"path"
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
	Task      *Task
	Build     *ChangeTaskBuild
	Deploy    []*ChangeTaskDeploy
	deployMap map[string]*ChangeTaskDeploy
}

func (p *ChangeTask) AddDeploy(bin *ChangeTaskDeploy) {

	index := fmt.Sprintf("%s-%s", bin.TaskId, bin.EnvId)
	if p.deployMap == nil {
		p.deployMap = make(map[string]*ChangeTaskDeploy)
	}

	if _, ok := (p.deployMap)[index]; ok {
		return
	}

	p.deployMap[index] = bin
	p.Deploy = append(p.Deploy, bin)
}

type ChangeTaskBuild struct {
	Modify   bool
	Md5      string
	Force    bool
	Type     int
	None     []*ChangeFile
	New      []*ChangeFile
	Update   []*ChangeFile
	Delete   []*ChangeFile
	FileList []*ChangeFile
	fileMap  map[string]*ChangeFile
}

type ChangeTaskDeploy struct {
	Modify bool
	Md5    string
	TaskId string
	EnvId  string
	Branch string
	Type   int
	Dist   string
	ModAt  int64
}

func (p *ChangeTaskDeploy) IsZip() bool {
	return strings.HasSuffix(p.Dist, enums.ZipSuffix)
}

func (p *ChangeTaskDeploy) DistName() string {
	return path.Base(p.Dist)
}

func (p *ChangeTaskBuild) SetMd5() {
	items := make([]string, 0)
	for _, v := range p.FileList {
		if v.Type == enums.ChangeTypeDelete {
			continue
		}
		items = append(items, v.Md5)
	}
	p.Md5 = secret.Md5([]byte(strings.Join(items, "")))
}

func (p *ChangeTaskBuild) Get(ident string) *ChangeFile {
	if v, ok := (p.fileMap)[ident]; ok {
		return v
	}
	return nil
}

func NewChangeTask(task *Task) *ChangeTask {

	changeTask := new(ChangeTask)
	changeTask.Task = task
	changeTask.Build = new(ChangeTaskBuild)
	changeTask.Build.fileMap = make(map[string]*ChangeFile)

	return changeTask
}

func (p *ChangeTaskBuild) Add(list *[]*ChangeFile, changeFile *ChangeFile) {

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

type WatchFiles struct {
	List   []string
	filter map[string]bool
}

func (p *WatchFiles) Has(path string) bool {
	if _, ok := p.filter[path]; ok {
		return true
	}
	return false
}

func (p *WatchFiles) Add(path string) {
	if p.filter == nil {
		p.filter = make(map[string]bool)
	}

	if _, ok := p.filter[path]; ok {
		return
	}

	p.filter[path] = true
	p.List = append(p.List, path)
}

func MakeChange() (change *Change, err error) {

	ctx, err := Prepare()
	if err != nil {
		return
	}

	change = new(Change)

	branch := Branch(ctx.Directory)

	for _, task := range ctx.Task {

		var taskRecord *TaskRecord
		taskRecord, err = FindTaskRecord(branch, task.Id)
		if err != nil {
			return
		}

		changeTask := NewChangeTask(task)

		if taskRecord == nil {
			taskRecord = NewTaskRecord(branch, task.Id)
			changeTask.Build.Type = enums.ChangeTypeNew
			change.Add(task.Id, changeTask)
		} else {
			changeTask.Build.Type = enums.ChangeTypeUpdate
			change.Add(task.Id, changeTask)
		}

		var files []string
		files, err = TaskGlobFiles(ctx, task)
		if err != nil {
			logger.Error("Match files error: ", err)
			return
		}

		for _, filePath := range files {
			var fileRecord *FileRecord
			fileRecord, err = FindFileRecord(branch, GetFileIdent(filePath))
			if err != nil {
				return
			}

			var modAt int64
			modAt, err = FileModAt(filePath)
			if err != nil {
				return
			}

			changeFile := NewChangeFile(filePath)
			changeFile.ModAt = modAt

			if fileRecord == nil {
				changeFile.Type = enums.ChangeTypeNew
				changeFile.Md5, err = FileMd5(filePath)
				if err != nil {
					return
				}
				changeTask.Build.Add(&changeTask.Build.New, changeFile)
			} else {
				changeFile.Md5, err = FileMd5(filePath)
				if err != nil {
					return
				}

				if modAt <= fileRecord.ModAt || changeFile.Md5 == fileRecord.Md5 {
					changeFile.Type = enums.ChangeTypeNone
					changeTask.Build.Add(&changeTask.Build.None, changeFile)

				} else {

					changeFile.Type = enums.ChangeTypeUpdate
					changeTask.Build.Add(&changeTask.Build.Update, changeFile)
				}
			}
		}

		changeTask.Build.Modify = true
		changeTask.Build.SetMd5()

		var taskFileRecords []*FileTaskRecord
		taskFileRecords, err = FindTaskFileRecords(branch, task.Id)
		if err != nil {
			return
		}

		for _, v := range taskFileRecords {
			changeFile := changeTask.Build.Get(v.FileIdent)
			if changeFile != nil {
				continue
			}
			var fileRecord *FileRecord
			fileRecord, err = FindFileRecord(branch, v.FileIdent)
			if err != nil {
				return
			}
			if fileRecord == nil {
				err = fmt.Errorf("file record \"%s\" not exist", v.FileIdent)
				return
			}
			changeFile = NewChangeFile(fileRecord.Path)
			changeFile.Type = enums.ChangeTypeDelete
			changeTask.Build.Add(&changeTask.Build.Delete, changeFile)
		}

		for _, build := range task.Build {

			if build.Bin != "" {

				taskBinDir := filepath.Join(storage.BinDir(), task.Id, build.Env, branch)

				if storage.Exist(taskBinDir) {

					var files []string
					files, err = storage.Files(taskBinDir, "")
					if err != nil {
						logger.Error("Get bin file error: ", err)
						return
					}

					var binRecord *BinRecord
					binRecord, err = FindBinRecord(task.Id, build.Env, branch)
					if err != nil {
						return
					}

					if len(files) > 1 {
						err = fmt.Errorf("multiple bin exist at \"%s\"", taskBinDir)
						logger.Error("Make change error: ", err)
						return
					}

					var taskBinFile string

					if len(files) == 1 {
						taskBinFile = files[0]
					}

					changeTaskDeploy := new(ChangeTaskDeploy)
					changeTaskDeploy.TaskId = task.Id
					changeTaskDeploy.EnvId = build.Env
					changeTaskDeploy.Branch = branch

					if taskBinFile == "" {

						if binRecord == nil {
							continue
						}

						changeTaskDeploy.Type = enums.ChangeTypeDelete

					} else {
						changeTaskDeploy.Dist = taskBinFile
						changeTaskDeploy.ModAt, err = FileModAt(taskBinFile)
						if err != nil {
							return
						}
						changeTaskDeploy.Md5, err = FileMd5(taskBinFile)
						if err != nil {
							return
						}
						changeTaskDeploy.Modify = true

						if binRecord == nil {
							changeTaskDeploy.Type = enums.ChangeTypeNew
						} else {
							changeTaskDeploy.Type = enums.ChangeTypeUpdate
							if binRecord.ModAt <= changeTaskDeploy.ModAt || binRecord.Md5 != changeTaskDeploy.Md5 {
								changeTaskDeploy.Modify = false
							}
						}
					}

					changeTask.AddDeploy(changeTaskDeploy)
				}

			}
		}
	}

	taskRecords, err := FindTaskRecords(branch)
	if err != nil {
		return
	}

	for _, v := range taskRecords {
		changeTask := change.Get(v.Id)
		if changeTask != nil {
			if changeTask.Build.Md5 == v.Md5 {
				changeTask.Build.Modify = false
			}
			continue
		}
		changeTask = NewChangeTask(&Task{
			Id: v.Id,
		})
		changeTask.Build.Type = enums.ChangeTypeDelete
		change.Add(v.Id, changeTask)
	}

	return
}

func FileModAt(filePath string) (modAt int64, err error) {
	var file os.FileInfo
	file, err = os.Stat(filePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Get file \"%s\" status error: ", filePath), err)
		return
	}
	modAt = file.ModTime().Unix()
	return
}

func TaskGlobFiles(ctx *Context, task *Task) (files []string, err error) {

	watchFiles := new(WatchFiles)

	for _, glob := range task.Watch {

		glob = filepath.Join(ctx.Directory, task.Directory, glob)

		var globFiles []string
		globFiles, err = filepath.Glob(glob)
		if err != nil {
			logger.Error("Match file error: ", err)
			return
		}

		for _, v := range globFiles {
			if strings.HasSuffix(v, enums.GoSuffix) {
				err = GoPackageFiles(watchFiles, v)
				if err != nil {
					return
				}
			}
			watchFiles.Add(v)
		}
	}

	files = watchFiles.List

	return
}
