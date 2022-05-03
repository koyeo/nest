package runner

import (
	"fmt"
	"github.com/gozelle/_color"
	"github.com/gozelle/_exec"
	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/utils/_tar"
	"github.com/koyeo/nest/utils/unit"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func NewServerRunner(conf *protocol.Config, task *TaskRunner, server *protocol.Server, key string) *ServerRunner {
	return &ServerRunner{
		conf:   conf,
		task:   task,
		key:    key,
		server: server,
		pool:   _exec.NewServerPool(),
	}
}

type ServerRunner struct {
	task   *TaskRunner
	conf   *protocol.Config
	key    string
	server *protocol.Server
	pool   *_exec.ServerPool
}

func (p *ServerRunner) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

func (p *ServerRunner) newExecServer() (*_exec.Server, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}
	if p.server.Port == 0 {
		p.server.Port = 22
	}
	if p.server.Password == "" && p.server.IdentityFile == "" {
		p.server.IdentityFile = "~/.ssh/id_rsa"
	}
	server := p.pool.New(&_exec.Server{
		Key:          p.key,
		Host:         p.server.Host,
		Port:         p.server.Port,
		User:         p.server.User,
		Password:     p.server.Password,
		IdentityFile: p.server.IdentityFile,
	})
	err := server.InitSFTP()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (p *ServerRunner) prepareTargetDir(target string) (dir string, err error) {
	server, err := p.newExecServer()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = server.SFTPClient().MkdirAll(dir)
			if err != nil {
				err = fmt.Errorf("make server target dir error: %s", err)
				return
			}
		}
	}()
	dir = target
	if !strings.HasSuffix(target, "/") {
		dir = filepath.Join(target, "../")
		return
	}
	return
}

func (p *ServerRunner) checkTargetPath(target string) (err error) {
	if !strings.HasPrefix(target, "/") && strings.HasPrefix(target, "~") {
		err = fmt.Errorf("invilad target path: '%s'", target)
		return
	}
	if strings.HasPrefix(target, "/") {
		items := strings.Split(strings.TrimPrefix(target, "/"), "/")
		if len(items) < 2 {
			err = fmt.Errorf("target path: %s too sort", target)
			return
		}
	}
	return
}

func (p *ServerRunner) Upload(source, target string) (err error) {

	err = p.checkTargetPath(target)
	if err != nil {
		return
	}

	ok, err := _fs.Exists(source)
	if err != nil {
		err = fmt.Errorf("detect upload source eroor: %s", source)
	}
	if !ok {
		err = fmt.Errorf("upload source: %s not exists", source)
		return
	}

	// prepare paths
	targetDir, err := p.prepareTargetDir(target)
	if err != nil {
		return
	}

	sourceName := path.Base(source)
	var targetName string
	if strings.HasSuffix(target, "/") {
		targetName = sourceName
	} else {
		targetName = path.Base(target)
	}
	bundleName := fmt.Sprintf("%s.tar.gz", sourceName)
	bundleLocalPath := fmt.Sprintf("%s/%s", NestTmpDir(), bundleName)
	defer func() {
		cleanNestTempDir()
	}()

	bundleRemoteTmpName := fmt.Sprintf("bundle-%s~", bundleName)
	bundleRemoteTmpPath := fmt.Sprintf("%s/%s", targetDir, bundleRemoteTmpName)
	var cmd string

	// compress source
	//cmd = fmt.Sprintf("tar -czf %s -C %s %s", bundleLocalPath, filepath.Join(source, "../"), path.Base(source))
	//fmt.Println(cmd)
	//_, err = _exec.NewRunner().AddCommand(cmd).CombinedOutput()
	//if err != nil {
	//	err = fmt.Errorf("compress source error: %s", err)
	//}
	// 通过 golang 的 tar 包压缩文件，解决向 linux 和 mac 的不兼容问题，此处还需要多加测试
	sourceFile, err := os.Open(source)
	if err != nil {
		err = fmt.Errorf("open source error: %s", err)
		return
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	err = _tar.Compress([]*os.File{sourceFile}, bundleLocalPath)
	if err != nil {
		err = fmt.Errorf("compress source error: %s", err)
		return
	}

	// upload
	bundleLocalFile, err := os.Open(bundleLocalPath)
	if err != nil {
		err = fmt.Errorf("open local bundle error: %s", err)
		return
	}
	defer func() {
		_ = bundleLocalFile.Close()
	}()
	bundleLocalInfo, err := bundleLocalFile.Stat()
	if err != nil {
		err = fmt.Errorf("stat local bundle error: %s", err)
		return
	}

	server, err := p.newExecServer()
	if err != nil {
		return
	}
	bundleRemoteTmpFile, err := server.SFTPClient().Create(bundleRemoteTmpPath)
	if err != nil {
		err = fmt.Errorf("create remote bundle error: %s", err)
		return
	}
	defer func() {
		_ = bundleRemoteTmpFile.Close()
		_ = server.SFTPClient().Remove(bundleRemoteTmpPath)
	}()

	// print upload progress
	size := 1 * 1024 * 1024
	buf := make([]byte, 1024*1024)
	total := unit.ByteSize(bundleLocalInfo.Size())
	uploaded := int64(0)
	p.printUpload(source, targetDir, targetName)
	for {
		n, _ := bundleLocalFile.Read(buf)
		if n == 0 {
			break
		}
		uploaded += int64(n)
		if n < size {
			_, err = bundleRemoteTmpFile.Write(buf[0:n])
		} else {
			_, err = bundleRemoteTmpFile.Write(buf)
		}
		if err != nil {
			err = fmt.Errorf("uplaod write remote bundle error:%s", err)
			return
		}
		fmt.Printf("\rTotal: %s Uploaded: %s", total, unit.ByteSize(uploaded))
	}
	fmt.Printf("\n")

	// recover upload bundle
	cmd = fmt.Sprintf("cd %s && rm -rf .nest && mkdir .nest && tar -xzf %s -C .nest", targetDir, bundleRemoteTmpName)
	//fmt.Println(cmd)
	err = p.CombinedExec(cmd)
	if err != nil {
		err = fmt.Errorf("recover upload bunle error: %s", err)
		return
	}
	_, err = server.SFTPClient().Stat(filepath.Join(targetDir, targetName))
	if err == nil {
		cmd = fmt.Sprintf("cd %s && rm -rf ./%s", targetDir, targetName)
		//fmt.Println(cmd)
		err = p.CombinedExec(cmd)
		if err != nil {
			err = fmt.Errorf("create target backup error: %s", err)
			return
		}
	}
	cmd = fmt.Sprintf("cd %s && mv .nest/%s ./%s && rm -rf .nest", targetDir, sourceName, targetName)
	//fmt.Println(cmd)
	err = p.CombinedExec(cmd)
	if err != nil {
		err = fmt.Errorf("mv upload bunle to target error: %s", err)
		return
	}

	return
}

func (p *ServerRunner) printUpload(source, targetDir, targetName string) {
	mapper := fmt.Sprintf("%s ===> %s", source, filepath.Join(targetDir, targetName))
	logger.Step(
		p.task.key,
		p.task.task.Comment,
		"🚀",
		_color.New(_color.FgCyan).Sprintf("[%s]", p.server.Name()),
		_color.New(_color.FgMagenta, _color.Bold).Sprintf("%s", mapper),
	)
}

func (p *ServerRunner) CombinedExec(command string) error {
	server, err := p.newExecServer()
	if err != nil {
		return err
	}
	return server.CombinedExec(command)
}

func (p *ServerRunner) PipeExec(command string) error {
	server, err := p.newExecServer()
	if err != nil {
		return err
	}
	return server.PipeExec(command)
}

func GetNestTempDir() string {
	return "./.nest/tmp"
}

func NestTmpDir() string {
	dir := GetNestTempDir()
	ok, err := _fs.Exists(dir)
	if err == nil && !ok {
		_ = _fs.MakeDir(dir)
	}
	return dir
}

func cleanNestTempDir() {
	_ = _fs.Remove(GetNestTempDir())
}
