package runner

import (
	"fmt"
	"github.com/gozelle/_exec"
	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/utils/unit"
	"os"
	"path"
	"path/filepath"
)

func NewServerRunner(conf *protocol.Config, server *protocol.Server, key string) *ServerRunner {
	return &ServerRunner{
		conf:   conf,
		key:    key,
		server: server,
		pool:   _exec.NewServerPool(),
	}
}

type ServerRunner struct {
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

func (p *ServerRunner) detectTargetDir(target string) (dir string, err error) {
	server, err := p.newExecServer()
	if err != nil {
		return
	}
	dir = target
	info, err := server.SFTPClient().Stat(dir)
	if err != nil {
		if err == os.ErrNotExist {
			err = server.SFTPClient().MkdirAll(dir)
			if err != nil {
				err = fmt.Errorf("make server target dir error: %s", err)
				return
			}
			return
		} else {
			err = fmt.Errorf("detect target error: %s", err)
			return
		}
	}

	if info.IsDir() {
		return
	}
	dir = filepath.Join(dir, "../")
	return
}

func (p *ServerRunner) Upload(source, target string) (err error) {

	ok, err := _fs.Exist(source)
	if err != nil {
		err = fmt.Errorf("detect upload source eroor: %s", source)
	}
	if !ok {
		err = fmt.Errorf("upload source: %s not exists", source)
		return
	}

	// prepare paths
	bundleName := fmt.Sprintf("%s.tar.gz", path.Base(source))
	bundleLocalPath := fmt.Sprintf("%s/%s", PrepareGetNestTempDir(), bundleName)

	targetDir, err := p.detectTargetDir(target)
	if err != nil {
		return
	}

	bundleRemoteTmpName := fmt.Sprintf("~bundle-%s", bundleName)
	bundleRemoteTmpPath := fmt.Sprintf("%s/%s", targetDir, bundleRemoteTmpName)

	// compress source
	_, err = _exec.NewRunner().AddCommand(fmt.Sprintf("tar -czf %s %s", bundleLocalPath, source)).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("compress source error: %s", err)
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
		fmt.Printf("\rtotal: %s uploaded: %s", total, unit.ByteSize(uploaded))
	}
	fmt.Printf("\n")

	// recover upload bundle
	err = server.SSHSession().Run(fmt.Sprintf("cd %s && tar -xzf %s", targetDir, bundleRemoteTmpName))
	if err != nil {
		err = fmt.Errorf("recover upload bunle error: %s", err)
		return
	}

	server.SSHSession().Run(fmt.Sprintf("cd %s && mv %s %s", targetDir, bundleRemoteTmpName, bundleName))
	if err != nil {
		err = fmt.Errorf("mv upload bunle to target: %s", err)
		return
	}

	// clean tmp dir
	cleanNestTempDir()

	return
}

func (p *ServerRunner) Exec(command string) error {
	server, err := p.newExecServer()
	if err != nil {
		return err
	}
	return server.SSHSession().Run(command)
}

func GetNestTempDir() string {
	return "./.nest/temp"
}

func PrepareGetNestTempDir() string {
	dir := GetNestTempDir()
	ok, err := _fs.Exist(dir)
	if err == nil && !ok {
		_ = _fs.MakeDir(dir)
	}
	return dir
}

func cleanNestTempDir() {
	_ = _fs.Remove(GetNestTempDir())
}
