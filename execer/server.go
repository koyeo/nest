package execer

import (
	"bufio"
	"fmt"
	"github.com/koyeo/nest/logger"
	"golang.org/x/crypto/ssh"
	"time"
)

func ServerRunScript(sshClient *ssh.Client, script string) (err error) {
	
	err = ServerRunCommand(sshClient, fmt.Sprintf(`echo '%s' | /bin/bash -s`, script))
	if err != nil {
		return
	}
	
	return
}

func ServerRunCommand(sshClient *ssh.Client, command string) (err error) {
	
	session, err := sshClient.NewSession()
	if err != nil {
		logger.Error("New remote ssh session error: ", err)
		return
	}
	
	defer func() {
		_ = session.Close()
	}()
	
	stderr, err := session.StderrPipe()
	if err != nil {
		logger.Error("[Run remote ssh command get stderr error]", err)
	}
	
	stdout, err := session.StdoutPipe()
	if err != nil {
		logger.Error("[Run remote ssh command get stdout error]", err)
	}
	
	out := make(chan string, 1048576)
	defer func() {
		for len(out) > 0 {
			time.Sleep(500 * time.Millisecond)
			close(out)
		}
	}()
	
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()
	
	go func() {
		for {
			m := <-out
			if m != "" {
				fmt.Println(m)
			}
		}
	}()
	
	err = session.Run(command)
	if err != nil {
		return
	}
	
	return
}

//
//func makeRemoteBinDir(sshClient *ssh.Client, ctx *Context, server *Server, projectId string) (path string, err error) {
//
//	session, err := sshClient.NewSession()
//	if err != nil {
//		logger.Error("New ssh session error: ", err)
//		return
//	}
//
//	defer func() {
//		_ = session.Close()
//	}()
//
//	if server.Workspace != "" {
//		path = filepath.Join(server.Workspace, enums.RemoteBinDir, projectId)
//		err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
//		if err != nil {
//			return
//		}
//		return
//	}
//
//	path = filepath.Join(enums.RemoteOptBinDir, projectId)
//	err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
//	if err == nil {
//		return
//	}
//
//	path = filepath.Join(fmt.Sprintf("~/%s/%s", enums.WorkspaceDir, enums.RemoteBinDir), projectId)
//	err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
//	if err != nil {
//		return
//	}
//	return
//}
//
//func makeRemoteScriptDir(sshClient *ssh.Client, sftpClient *sftp.Client, ctx *Context, server *Server, projectId string) (path string, err error) {
//
//	session, err := sshClient.NewSession()
//	if err != nil {
//		logger.Error("New ssh session error: ", err)
//		return
//	}
//
//	defer func() {
//		_ = session.Close()
//	}()
//	var exist bool
//	if server.Workspace != "" {
//		path = filepath.Join(server.Workspace, enums.RemoteScriptDir, projectId)
//		exist, err = remoteExist(sftpClient, path)
//		if err != nil {
//			return
//		}
//		if !exist {
//			err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
//			if err != nil {
//				return
//			}
//		}
//		return
//	}
//
//	path = filepath.Join(enums.RemoteOptScriptDir, projectId)
//	exist, err = remoteExist(sftpClient, path)
//	if err != nil {
//		return
//	}
//	if !exist {
//		err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
//		if err == nil {
//			return
//		}
//	}
//
//	path = filepath.Join(fmt.Sprintf("~/%s/%s", enums.WorkspaceDir, enums.RemoteScriptDir), projectId)
//	exist, err = remoteExist(sftpClient, path)
//	if err != nil {
//		return
//	}
//	if !exist {
//		err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
//		if err != nil {
//			return
//		}
//	}
//
//	return
//}
//
//func remoteExist(sftpClient *sftp.Client, path string) (exist bool, err error) {
//
//	_, err = sftpClient.Stat(path)
//	if err == os.ErrNotExist {
//		return
//	} else if err != nil {
//		logger.Error(fmt.Sprintf("Stat remote \"%s\" path error", path), err)
//		return
//	}
//
//	exist = true
//
//	return
//}
//
//func PrepareScript(ctx *Context, script *Script) (content string, err error) {
//
//	if script.File == "" {
//		err = fmt.Errorf("script \"%s\" file is enpty", script.Id)
//		logger.Error("Prepare script error:", err)
//		return
//	}
//
//	originScriptFile := storage.ParsePath(ctx.Directory, script.File)
//
//	if !storage.Exist(originScriptFile) {
//		err = fmt.Errorf("script file \"%s\" not exist", storage.Abs(originScriptFile))
//		logger.Error("Prepare script error:", err)
//		return
//	}
//
//	data, err := storage.Read(originScriptFile)
//	if err != nil {
//		logger.Error("Prepare script file error: ", err)
//		return
//	}
//
//	for k, v := range script.Vars {
//		reg := regexp.MustCompile(`\{\{\s*\` + k + `\s*\}\}`)
//		data = reg.ReplaceAll(data, []byte(v))
//	}
//	content = string(data)
//	return
//}
//
//func uploadScript(sshClient *ssh.Client, sftpClient *sftp.Client, ctx *Context, server *Server, script *Script) (remoteScriptFile string, err error) {
//
//	if script.File == "" {
//		return
//	}
//
//	originScriptFile := storage.ParsePath(ctx.Directory, script.File)
//
//	if !storage.Exist(originScriptFile) {
//		err = fmt.Errorf("local script \"%s\" not exist", storage.Abs(originScriptFile))
//		logger.Error("Upload script error:", err)
//		return
//	}
//
//	remoteScriptDir, err := makeRemoteScriptDir(sshClient, sftpClient, ctx, server, ctx.Id)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Make remote script dir \"%s\" error: ", remoteScriptDir), err)
//		return
//	}
//
//	remoteScriptFile = filepath.Join(remoteScriptDir, path.Base(script.File))
//
//	if exist := ctx.AddUploadedScript(script.Id); exist {
//		return
//	}
//
//	info, err := sftpClient.Stat(remoteScriptFile)
//	if err == nil {
//		if info.IsDir() {
//			err = fmt.Errorf("script path \"%s\" is a dir", remoteScriptFile)
//			logger.Error("Make remote script file error: ", err)
//			return
//		} else {
//			err = SSHPipeRunCommand(sshClient, fmt.Sprintf("rm %s", remoteScriptFile), ctx.cli.Bool(enums.HideCommandFlag))
//			if err != nil {
//				logger.Error("Stat remote script file error: ", err)
//				return
//			}
//		}
//	}
//
//	file, err := sftpClient.Create(remoteScriptFile)
//	if err != nil {
//		logger.Error(fmt.Sprintf("Create remote script file \"%s\" error: ", remoteScriptFile), err)
//		return
//	}
//
//	data, err := storage.Read(originScriptFile)
//	if err != nil {
//		logger.Error("Read script file error: ", err)
//		return
//	}
//
//	n, err := file.Write(data)
//	if err != nil {
//		logger.Error("Write remote script file error: ", err)
//		return
//	}
//
//	if n == 0 {
//		err = fmt.Errorf("script file \"%s\" is empty", originScriptFile)
//		logger.Error("Create scirpt file error: ", err)
//		return
//	}
//
//	//err = SSHPipeRunCommand(sshClient, "chmod +x "+remoteScriptFile)
//	//if err != nil {
//	//	return
//	//}
//
//	return
//}
