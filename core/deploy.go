package core

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"nest/enums"
	"nest/logger"
	"nest/storage"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
		err = SSHPipeRunCommand(sshClient, v, ctx.cli.Bool(enums.PrintScriptFlag))
		if err != nil {
			return
		}
	}

	if len(deploy.BeforeScript) > 0 {
		fmt.Println("Exec before script:")
	}
	//var remoteScriptFile string
	for _, v := range deploy.BeforeScript {
		err = RunScript(sshClient, ctx, v, ctx.cli.Bool(enums.PrintScriptFlag))
		if err != nil {
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
		err = SSHPipeRunCommand(sshClient, v, ctx.cli.Bool(enums.PrintScriptFlag))
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
		err = SSHPipeRunCommand(sshClient, v, ctx.cli.Bool(enums.PrintScriptFlag))
		if err != nil {
			return
		}
	}

	if len(deploy.AfterScript) != 0 {
		fmt.Println("Exec after script:")
	}
	for _, v := range deploy.AfterScript {
		err = RunScript(sshClient, ctx, v, ctx.cli.Bool(enums.PrintScriptFlag))
		if err != nil {
			return
		}
	}

	return
}

func ExecStart(ctx *Context, task *Task, server *Server, start *DaemonStart) (err error) {

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

	if len(start.BeforeCommand) != 0 {
		fmt.Println("Exec before command:")
	}
	for _, v := range start.BeforeCommand {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err = SSHPipeRunCommand(sshClient, v, ctx.cli.Bool(enums.HideCommandFlag))
		if err != nil {
			return
		}
	}

	if len(start.BeforeScript) > 0 {
		fmt.Println("Exec before script:")
	}
	//var remoteScriptFile string
	for _, v := range start.BeforeScript {
		err = RunScript(sshClient, ctx, v, ctx.cli.Bool(enums.PrintScriptFlag))
		if err != nil {
			return
		}
	}

	if len(start.Command) > 0 {
		fmt.Println("Exec command:", len(start.Command))
	}
	for _, v := range start.Command {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err = SSHPipeRunCommand(sshClient, v, ctx.cli.Bool(enums.HideCommandFlag))
		if err != nil {
			return
		}
	}

	if len(start.AfterCommand) != 0 {
		fmt.Println("Exec after command:")
	}
	for _, v := range start.AfterCommand {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err = SSHPipeRunCommand(sshClient, v, ctx.cli.Bool(enums.HideCommandFlag))
		if err != nil {
			return
		}
	}

	if len(start.AfterScript) != 0 {
		fmt.Println("Exec after script:")
	}
	for _, v := range start.AfterScript {
		err = RunScript(sshClient, ctx, v, ctx.cli.Bool(enums.PrintScriptFlag))
		if err != nil {
			return
		}
	}

	return
}

func home() (path string, err error) {
	path, err = Exec("", "echo ~")
	if err != nil {
		logger.Error("Get home path error", err)
		return
	}

	return
}

func publicKey(path string) (auth ssh.AuthMethod, err error) {

	u, err := home()
	if err != nil {
		return
	}

	if strings.HasPrefix(path, "~") {
		path = filepath.Join(u, strings.TrimPrefix(path, "~"))
	}

	key, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error("Read identity error:", err)
		return
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		logger.Error("Read identity error:", err)
		return
	}
	auth = ssh.PublicKeys(signer)

	return
}

func SSHClient(server *Server) (client *ssh.Client, err error) {

	var auth []ssh.AuthMethod

	if server.SSH.Password != "" {
		auth = append(auth, ssh.Password(server.SSH.Password))
	}

	if server.SSH.Identity != "" {
		var ident ssh.AuthMethod
		ident, err = publicKey(server.SSH.Identity)
		if err != nil {
			return
		}
		auth = append(auth, ident)
	}

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

	err = uploadBin(sshClient, sftpClient, ctx, deploy, changeDeploy)
	if err != nil {
		return
	}

	return
}

func uploadBin(sshClient *ssh.Client, sftpClient *sftp.Client, ctx *Context, task *Task, env *Env, deploy *Deploy, changeDeploy *ChangeTaskDeploy) (err error) {

	binPath := deploy.Bin.Path

	info, err := sftpClient.Stat(binPath)

	if err == nil {
		if !info.IsDir() {
			err = fmt.Errorf("bin path \"%s\" is a file", binPath)
			logger.Error("Create bin path error: ", err)
			return
		} else {
			err = SSHPipeRunCommand(sshClient, fmt.Sprintf("rm -rf %s/*", binPath), ctx.cli.Bool(enums.HideCommandFlag))
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

	err = SSHPipeRunCommand(sshClient, fmt.Sprintf("cd %s && unzip %s", binPath, binName), ctx.cli.Bool(enums.HideCommandFlag))
	if err != nil {
		return
	}

	err = SSHPipeRunCommand(sshClient, fmt.Sprintf("rm %s", binFile), ctx.cli.Bool(enums.HideCommandFlag))
	if err != nil {
		return
	}

	return
}

func makeRemoteBinDir(sshClient *ssh.Client, ctx *Context, server *Server, projectId string) (path string, err error) {

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
		err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
		if err != nil {
			return
		}
		return
	}

	path = filepath.Join(enums.RemoteOptBinDir, projectId)
	err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
	if err == nil {
		return
	}

	path = filepath.Join(fmt.Sprintf("~/%s/%s", enums.WorkspaceDir, enums.RemoteBinDir), projectId)
	err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
	if err != nil {
		return
	}
	return
}

func makeRemoteScriptDir(sshClient *ssh.Client, sftpClient *sftp.Client, ctx *Context, server *Server, projectId string) (path string, err error) {

	session, err := sshClient.NewSession()
	if err != nil {
		logger.Error("New ssh session error: ", err)
		return
	}

	defer func() {
		_ = session.Close()
	}()
	var exist bool
	if server.Workspace != "" {
		path = filepath.Join(server.Workspace, enums.RemoteScriptDir, projectId)
		exist, err = remoteExist(sftpClient, path)
		if err != nil {
			return
		}
		if !exist {
			err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
			if err != nil {
				return
			}
		}
		return
	}

	path = filepath.Join(enums.RemoteOptScriptDir, projectId)
	exist, err = remoteExist(sftpClient, path)
	if err != nil {
		return
	}
	if !exist {
		err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
		if err == nil {
			return
		}
	}

	path = filepath.Join(fmt.Sprintf("~/%s/%s", enums.WorkspaceDir, enums.RemoteScriptDir), projectId)
	exist, err = remoteExist(sftpClient, path)
	if err != nil {
		return
	}
	if !exist {
		err = SSHPipeRunCommand(sshClient, "make -p "+path, ctx.cli.Bool(enums.HideCommandFlag))
		if err != nil {
			return
		}
	}

	return
}

func remoteExist(sftpClient *sftp.Client, path string) (exist bool, err error) {

	_, err = sftpClient.Stat(path)
	if err == os.ErrNotExist {
		return
	} else if err != nil {
		logger.Error(fmt.Sprintf("Stat remote \"%s\" path error", path), err)
		return
	}

	exist = true

	return
}

func PrepareScript(ctx *Context, script *Script) (content string, err error) {

	if script.File == "" {
		err = fmt.Errorf("script \"%s\" file is enpty", script.Id)
		logger.Error("Prepare script error:", err)
		return
	}

	originScriptFile := storage.ParsePath(ctx.Directory, script.File)

	if !storage.Exist(originScriptFile) {
		err = fmt.Errorf("script file \"%s\" not exist", storage.Abs(originScriptFile))
		logger.Error("Prepare script error:", err)
		return
	}

	data, err := storage.Read(originScriptFile)
	if err != nil {
		logger.Error("Prepare script file error: ", err)
		return
	}

	for k, v := range script.Vars {
		reg := regexp.MustCompile(`\{\{\s*\` + k + `\s*\}\}`)
		data = reg.ReplaceAll(data, []byte(v))
	}
	content = string(data)
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

	remoteScriptDir, err := makeRemoteScriptDir(sshClient, sftpClient, ctx, server, ctx.Id)
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
			err = SSHPipeRunCommand(sshClient, fmt.Sprintf("rm %s", remoteScriptFile), ctx.cli.Bool(enums.HideCommandFlag))
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

	//err = SSHPipeRunCommand(sshClient, "chmod +x "+remoteScriptFile)
	//if err != nil {
	//	return
	//}

	return
}
