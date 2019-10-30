package core

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"nest/enums"
	"nest/logger"
	"nest/storage"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func ExecDeploy(ctx *Context, task *Task, server *Server, deploy *Deploy, changeDeploy *ChangeTaskDeploy) (err error) {

	if deploy.Bin == nil {
		err = fmt.Errorf("deploy \"%s\" bin is empty", deploy.Id)
		logger.Error("Deploy error: ", err)
		return
	}

	if deploy.Bin.Source != enums.BinSourceBuild && deploy.Bin.Source != enums.BinSourceUrl {
		err = fmt.Errorf("deploy \"%s\" invalid bin source \"%s\"", deploy.Id, deploy.Bin.Source)
		logger.Error("Deploy error: ", err)
		return
	}

	sshClient, err := SSHClient(server)
	if err != nil {
		return
	}
	defer func() {
		_ = sshClient.Close()
	}()

	sftpClient, err := SFTPClient(sshClient)
	if err != nil {
		return
	}
	defer func() {
		_ = sftpClient.Close()
	}()

	switch deploy.Bin.Source {
	case enums.BinSourceBuild:
		err = deploySourceBuild(sshClient, sftpClient, ctx, task, server, deploy, changeDeploy)
		if err != nil {
			return
		}
	}

	if len(deploy.BeforeCommand) != 0 {
		fmt.Println("Exec before command:")
	}
	for _, v := range deploy.BeforeCommand {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err = SSHPipeExec(sshClient, v)
		if err != nil {
			return
		}
	}

	if len(deploy.BeforeScript) > 0 {
		fmt.Println("Exec before script:")
		fmt.Printf("%+v\n", deploy.BeforeScript[0])
	}
	var remoteScriptFile string
	for _, v := range deploy.BeforeScript {
		remoteScriptFile, err = uploadScript(sshClient, sftpClient, ctx, server, v)
		if err != nil {
			return
		}
		err = SSHPipeExec(sshClient, "/bin/bash "+remoteScriptFile)
		if err != nil {
			logger.Error(fmt.Sprintf("Run remote berfore script \"%s\" error: ", remoteScriptFile), err)
			return
		}
	}

	if len(deploy.Command) > 0 {
		fmt.Println("Exec command:", len(deploy.Command))
	}
	for _, v := range deploy.Command {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err = SSHPipeExec(sshClient, v)
		if err != nil {
			return
		}
	}

	if len(deploy.AfterCommand) != 0 {
		fmt.Println("Exec after command:")
	}
	for _, v := range deploy.AfterCommand {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err = SSHPipeExec(sshClient, v)
		if err != nil {
			return
		}
	}

	if len(deploy.AfterScript) != 0 {
		fmt.Println("Exec after script:")
	}
	for _, v := range deploy.AfterScript {
		remoteScriptFile, err = uploadScript(sshClient, sftpClient, ctx, server, v)
		if err != nil {
			return
		}
		err = SSHPipeExec(sshClient, "/bin/bash "+remoteScriptFile)
		if err != nil {
			logger.Error(fmt.Sprintf("Run remote after script \"%s\" error: ", remoteScriptFile), err)
			return
		}
	}

	return
}

func SSHClient(server *Server) (client *ssh.Client, err error) {

	var auth []ssh.AuthMethod
	auth = append(auth, ssh.Password(server.SSH.Password))
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.SSH.Ip, server.SSH.Port), &ssh.ClientConfig{
		User:            server.SSH.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})

	if err != nil {
		logger.Error("Connect server via ssh error: ", err)
		return
	}

	return
}

func SFTPClient(sshClient *ssh.Client) (client *sftp.Client, err error) {

	client, err = sftp.NewClient(sshClient)
	if err != nil {
		logger.Error("Connect server via sftp error: ", err)
		return
	}
	return
}

func deploySourceBuild(sshClient *ssh.Client, sftpClient *sftp.Client, ctx *Context, task *Task, server *Server, deploy *Deploy, changeDeploy *ChangeTaskDeploy) (err error) {

	err = uploadBin(sshClient, sftpClient, deploy, changeDeploy)
	if err != nil {
		return
	}

	return
}

func uploadBin(sshClient *ssh.Client, sftpClient *sftp.Client, deploy *Deploy, changeDeploy *ChangeTaskDeploy) (err error) {

	binPath := deploy.Bin.Path

	info, err := sftpClient.Stat(binPath)

	if err == nil {
		if !info.IsDir() {
			err = fmt.Errorf("bin path \"%s\" is a file", binPath)
			logger.Error("Create bin path error: ", err)
			return
		} else {
			err = SSHPipeExec(sshClient, fmt.Sprintf("rm -rf %s/*", binPath))
			if err != nil {
				return
			}
		}
	} else {
		if os.IsNotExist(err) {
			err = sftpClient.MkdirAll(binPath)
			if err != nil {
				logger.Error(fmt.Sprintf("Create bin path \"%s\" error: ", binPath), err)
				return
			}
		} else {
			logger.Error("GetTask path info error: ", err)
			return
		}
	}

	binName := changeDeploy.DistName()
	if !changeDeploy.IsZip() {
		err = fmt.Errorf("bin file \"%s\" is not zip file", binName)
		logger.Error("Create bin file error: ", err)
	}

	binFile := filepath.Join(binPath, binName)
	file, err := sftpClient.Create(binFile)
	if err != nil {
		logger.Error("Create bin file error: ", err)
		return
	}

	data, err := storage.Read(changeDeploy.Dist)
	if err != nil {
		logger.Error("Read bin file error: ", err)
		return
	}

	n, err := file.Write(data)
	if err != nil {
		logger.Error("Write bin file error: ", err)
		return
	}

	if n == 0 {
		err = fmt.Errorf("bin file \"%s\" is empty", binFile)
		logger.Error("Create bin file error: ", err)
		return
	}

	err = SSHPipeExec(sshClient, fmt.Sprintf("cd %s && unzip %s", binPath, binName))
	if err != nil {
		return
	}

	err = SSHPipeExec(sshClient, fmt.Sprintf("rm %s", binFile))
	if err != nil {
		return
	}

	return
}

func makeRemoteBinDir(sshClient *ssh.Client, server *Server, projectId string) (path string, err error) {

	session, err := sshClient.NewSession()
	if err != nil {
		logger.Error("New ssh session error: ", err)
		return
	}

	defer func() {
		_ = session.Close()
	}()

	if server.Workspace != "" {
		path = filepath.Join(server.Workspace, enums.RemoteBinDir, projectId)
		err = SSHPipeExec(sshClient, "make -p "+path)
		if err != nil {
			return
		}
		return
	}

	path = filepath.Join(enums.RemoteOptBinDir, projectId)
	err = SSHPipeExec(sshClient, "make -p "+path)
	if err == nil {
		return
	}

	path = filepath.Join(fmt.Sprintf("~/%s/%s", enums.WorkspaceDir, enums.RemoteBinDir), projectId)
	err = SSHPipeExec(sshClient, "make -p "+path)
	if err != nil {
		return
	}
	return
}

func makeRemoteScriptDir(sshClient *ssh.Client, server *Server, projectId string) (path string, err error) {

	session, err := sshClient.NewSession()
	if err != nil {
		logger.Error("New ssh session error: ", err)
		return
	}

	defer func() {
		_ = session.Close()
	}()

	if server.Workspace != "" {
		path = filepath.Join(server.Workspace, enums.RemoteScriptDir, projectId)
		err = SSHPipeExec(sshClient, "make -p "+path)
		if err != nil {
			return
		}
		return
	}

	path = filepath.Join(enums.RemoteOptScriptDir, projectId)
	err = SSHPipeExec(sshClient, "make -p "+path)
	if err == nil {
		return
	}

	path = filepath.Join(fmt.Sprintf("~/%s/%s", enums.WorkspaceDir, enums.RemoteScriptDir), projectId)
	err = SSHPipeExec(sshClient, "make -p "+path)
	if err != nil {
		return
	}

	return
}

func uploadScript(sshClient *ssh.Client, sftpClient *sftp.Client, ctx *Context, server *Server, script *Script) (remoteScriptFile string, err error) {

	if script.File == "" {
		return
	}

	originScriptFile := storage.ParsePath(ctx.Directory, script.File)

	if !storage.Exist(originScriptFile) {
		err = fmt.Errorf("local script \"%s\" not exist", storage.Abs(originScriptFile))
		logger.Error("Upload script error:", err)
		return
	}

	remoteScriptDir, err := makeRemoteScriptDir(sshClient, server, ctx.Id)
	if err != nil {
		logger.Error(fmt.Sprintf("Make remote script dir \"%s\" error: ", remoteScriptDir), err)
		return
	}

	remoteScriptFile = filepath.Join(remoteScriptDir, path.Base(script.File))

	if exist := ctx.AddUploadedScript(script.Id); exist {
		return
	}

	info, err := sftpClient.Stat(remoteScriptFile)
	if err == nil {
		if info.IsDir() {
			err = fmt.Errorf("script path \"%s\" is a dir", remoteScriptFile)
			logger.Error("Make remote script file error: ", err)
			return
		} else {
			err = SSHPipeExec(sshClient, fmt.Sprintf("rm %s", remoteScriptFile))
			if err != nil {
				logger.Error("Stat remote script file error: ", err)
				return
			}
		}
	}

	file, err := sftpClient.Create(remoteScriptFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Create remote script file \"%s\" error: ", remoteScriptFile), err)
		return
	}

	data, err := storage.Read(originScriptFile)
	if err != nil {
		logger.Error("Read script file error: ", err)
		return
	}

	n, err := file.Write(data)
	if err != nil {
		logger.Error("Write remote script file error: ", err)
		return
	}

	if n == 0 {
		err = fmt.Errorf("script file \"%s\" is empty", originScriptFile)
		logger.Error("Create scirpt file error: ", err)
		return
	}

	//err = SSHPipeExec(sshClient, "chmod +x "+remoteScriptFile)
	//if err != nil {
	//	return
	//}

	return
}
