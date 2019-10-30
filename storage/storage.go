package storage

import (
	"fmt"
	"io/ioutil"
	"nest/enums"
	"nest/logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ParsePath(root string, path ...string) string {
	for k, v := range path {
		if strings.HasPrefix(v, enums.RelativePathPrefix) {
			if k < len(path)-1 {
				return filepath.Join(append([]string{v}, path[k+1:]...)...)
			} else {
				return v
			}
		}
	}
	return filepath.Join(append([]string{root}, path...)...)
}

func WorkspaceDir(path ...string) string {
	return filepath.Join("./", enums.WorkspaceDir, filepath.Join(path...))
}

func DataFile() string {
	return WorkspaceDir(enums.DataDir, enums.DataFile)
}

func DataDir() string {
	return WorkspaceDir(enums.DataDir)
}

func BinDir() string {
	return WorkspaceDir(enums.BinDir)
}

func BinFile(file string) string {
	return WorkspaceDir(enums.BinDir, file)
}

func Read(path string) (data []byte, err error) {

	data, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return
}

func Remove(path string) (err error) {
	err = os.RemoveAll(path)
	return
}

func Exist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		logger.Error("Check path exist error: ", err)
	}
	return true
}

func Files(path string, suffix ...string) (files []string, err error) {

	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	sep := string(os.PathSeparator)

	for _, fi := range dir {
		if fi.IsDir() {
			var dirFiles []string
			dirFiles, err = Files(filepath.Join(path, sep, fi.Name()), suffix...)
			if err != nil {
				return
			}
			files = append(files, dirFiles...)
		} else {

			if hasSuffix(suffix, fi.Name()) {
				files = append(files, filepath.Join(path, sep, fi.Name()))
			}
		}
	}

	return
}

func hasSuffix(suffix []string, fileName string) bool {
	for _, v := range suffix {
		if strings.HasSuffix(fileName, v) {
			return true
		}
	}

	return false
}

func Root() string {

	root, err := os.Getwd()
	if err != nil {
		logger.Error("Getwd error: ", err)
		os.Exit(1)
	}

	return root
}

func MakeDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("mkdir -p %s", path))
		err = cmd.Run()
		if err != nil {
			logger.Error("Make dir error: ", err)
		}
	}
}

func Abs(path ...string) string {
	return filepath.Join(Root(), filepath.Join(path[:]...))
}

func Write(path string, content []byte) (err error) {
	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		return
	}
	return
}

func Relative(path string) string {
	wd, _ := os.Getwd()
	path = strings.TrimPrefix(path, wd+"/")
	return path
}
