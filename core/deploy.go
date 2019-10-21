package core

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"nest/enums"
	"nest/logger"
	"nest/storage"
	"os"
	"path/filepath"
	"time"
)

// 1. ssh 连接
// 2. sftp 连接
// 3. 更新 Bin
// 4. 更新 supervisor 配置文件
// 5. 检查 supervisor 状态
// 6. 执行前置脚本
// 7.
// 6. 启动或重启程序
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

	return
}

func SSHClient(server *Server) (client *ssh.Client, err error) {

	var auth []ssh.AuthMethod
	auth = append(auth, ssh.Password(server.SSH.Password))
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Ip, server.SSH.Port), &ssh.ClientConfig{
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
			logger.Error("Get path info error: ", err)
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
