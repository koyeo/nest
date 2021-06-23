package core

import (
	"fmt"
	"github.com/koyeo/nest/constant"
	"github.com/koyeo/nest/execer"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/storage"
	"github.com/koyeo/snowflake"
	"github.com/pkg/sftp"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type TaskManager struct {
	_list []*Task
	_map  map[string]*Task
}

func (p *TaskManager) Add(item *Task) error {
	
	if item.Name == "" {
		return fmt.Errorf("task with empty name")
	}
	
	if p._map == nil {
		p._map = map[string]*Task{}
	}
	
	if _, ok := p._map[item.Name]; ok {
		return fmt.Errorf("duplicated task: %s", item.Name)
	}
	
	p._map[item.Name] = item
	p._list = append(p._list, item)
	
	return nil
}

func (p *TaskManager) Get(name string) *Task {
	if p._map == nil {
		return nil
	}
	return p._map[name]
}

func (p *TaskManager) List() []*Task {
	return p._list
}

type Task struct {
	Name                 string
	BuildCommand         Statement
	BuildScriptFile      string
	DeployPath           Statement
	DeployCommand        Statement
	DeploySource         Statement
	DeployScript         []byte
	DeploySupervisorConf []byte
	DeployServer         []*Server
	Flow                 []*Task
}

func (p *Task) vars() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *Task) Run(c *cli.Context) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	if len(p.Flow) > 0 {
		log.Println(chalk.Green.Color("[run task flow]"), p.Name)
		for _, v := range p.Flow {
			err = v.build(pwd, p.makeVars(v, pwd))
			if err != nil {
				return
			}
		}
		
		for _, v := range p.Flow {
			err = v.deploy(p.makeVars(v, pwd))
			if err != nil {
				return
			}
		}
	} else {
		return p.run()
	}
	
	// clean
	_ = storage.Remove(filepath.Join(pwd, constant.NEST_WORKSPACE))
	
	return
}

func (p *Task) run() (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	vars := p.makeVars(p, pwd)
	log.Println(chalk.Green.Color("[Run task]"), p.Name)
	err = p.build(pwd, vars)
	if err != nil {
		return
	}
	
	err = p.deploy(vars)
	if err != nil {
		return
	}
	
	logger.Successf("[Done]")
	
	return
}

func (p *Task) makeVars(task *Task, pwd string) map[string]string {
	vars := map[string]string{}
	vars[constant.NAME] = task.Name
	vars[constant.SELF] = task.Name
	vars[constant.WORKSPACE] = filepath.Join(pwd, constant.NEST_WORKSPACE, task.Name)
	return vars
}

func (p *Task) build(pwd string, vars map[string]string) (err error) {
	
	defer func() {
		if err != nil {
			err = nil
			os.Exit(1)
		}
	}()
	
	printBuild := false
	
	if p.BuildCommand != "" {
		printBuild = true
		log.Println(chalk.Green.Color("[build start]"))
		err = p.execBuildCommand(pwd, vars)
		if err != nil {
			return
		}
	}
	
	if p.BuildScriptFile != "" {
		if !printBuild {
			printBuild = true
			log.Println(chalk.Green.Color("[build start]"))
		}
		err = p.execBuildScriptFile(pwd, vars)
		if err != nil {
			return
		}
	}
	if printBuild {
		log.Println(chalk.Green.Color("[build end]"))
	}
	return
}

func (p *Task) execBuildCommand(pwd string, vars map[string]string) (err error) {
	command, err := p.BuildCommand.Render(vars)
	if err != nil {
		return
	}
	workspace := vars[constant.WORKSPACE]
	if !storage.Exist(workspace) {
		storage.MakeDir(workspace)
	}
	
	log.Println(chalk.Green.Color("[exec command]"), command)
	
	err = execer.RunCommand("", pwd, command)
	if err != nil {
		return
	}
	return
}

func (p *Task) execBuildScriptFile(pwd string, vars map[string]string) (err error) {
	
	bs, err := storage.Read(p.BuildScriptFile)
	if err != nil {
		err = fmt.Errorf("task '%s' read build_script_file='%s' error: %s", p.Name, p.BuildScriptFile, err)
		return
	}
	
	content, err := Statement(bs).Render(vars)
	if err != nil {
		return
	}
	
	gen, err := snowflake.DefaultGenerator(1)
	if err != nil {
		err = fmt.Errorf("new snowflake generate error: %s", err)
		return
	}
	
	id, err := gen.Generate()
	if err != nil {
		err = fmt.Errorf("snowflake generate id error: %s", err)
		return
	}
	
	tmpName := fmt.Sprintf("%s-%s", filepath.Base(p.BuildScriptFile), id.String())
	tmpScript := filepath.Join(vars[constant.WORKSPACE], tmpName)
	err = storage.Write(tmpScript, []byte(content))
	if err != nil {
		err = fmt.Errorf("write tmp script error: %s", err)
		return
	}
	
	defer func() {
		_ = storage.Remove(tmpScript)
	}()
	
	log.Println(chalk.Green.Color("[exec script]"), p.BuildScriptFile)
	
	err = execer.RunScript("", pwd, tmpScript)
	if err != nil {
		return
	}
	
	return
}

func (p *Task) deploy(vars map[string]string) (err error) {
	
	defer func() {
		if err != nil {
			err = nil
			os.Exit(1)
		}
	}()
	
	for _, v := range p.DeployServer {
		err = p.deployServer(v, vars)
		if err != nil {
			return
		}
	}
	return
}

func (p *Task) deployServer(server *Server, vars map[string]string) (err error) {
	
	log.Println(chalk.Green.Color("[deploy start]"), fmt.Sprintf("%s (%s)", chalk.Yellow.Color(server.Name), server.Host))
	defer func() {
		if err != nil {
			return
		}
		log.Println(chalk.Green.Color("[deploy end]"), fmt.Sprintf("%s (%s)", chalk.Yellow.Color(server.Name), server.Host))
	}()
	
	sshClient, err := p.newSSHClient(server)
	if err != nil {
		return
	}
	defer func() {
		_ = sshClient.Close()
	}()
	
	sftpClient, err := p.newSFTPClient(sshClient)
	if err != nil {
		return
	}
	defer func() {
		_ = sftpClient.Close()
	}()
	
	err = p.uploadTarget(sshClient, sftpClient, server, vars)
	if err != nil {
		logger.Error("[deploy error]", err)
		return
	}
	
	if p.DeployCommand != "" {
		var cmd string
		cmd, err = p.DeployCommand.Render(vars)
		if err != nil {
			return
		}
		err = execer.ServerRunCommand(sshClient, cmd)
		if err != nil {
			return
		}
	}
	
	if len(p.DeployScript) > 0 {
		err = execer.ServerRunScript(sshClient, string(p.DeployScript))
		if err != nil {
			return
		}
	}
	
	return
}

func (p *Task) newSSHPublicKey(path string) (auth ssh.AuthMethod, err error) {
	
	home, err := execer.HomePath()
	if err != nil {
		return
	}
	
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	
	key, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}
	auth = ssh.PublicKeys(signer)
	
	return
}

func (p *Task) newSSHClient(server *Server) (client *ssh.Client, err error) {
	
	var auth []ssh.AuthMethod
	
	if server.Password != "" {
		auth = append(auth, ssh.Password(server.Password))
	}
	
	if server.IdentityFile != "" {
		var ident ssh.AuthMethod
		ident, err = p.newSSHPublicKey(server.IdentityFile)
		if err != nil {
			return
		}
		auth = append(auth, ident)
	}
	
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port), &ssh.ClientConfig{
		User:            server.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	
	if err != nil {
		return
	}
	
	return
}

func (p *Task) newSFTPClient(sshClient *ssh.Client) (client *sftp.Client, err error) {
	
	client, err = sftp.NewClient(sshClient)
	if err != nil {
		return
	}
	return
}

func (p *Task) uploadTarget(sshClient *ssh.Client, sftpClient *sftp.Client, server *Server, vars map[string]string) (err error) {
	
	deploySource, err := p.DeploySource.Render(vars)
	if err != nil {
		return
	}
	
	deployPath, err := p.DeployPath.Render(vars)
	if err != nil {
		return
	}
	
	reg := regexp.MustCompile(`/.+`)
	if !reg.MatchString(deployPath) {
		err = fmt.Errorf("iggleal deployt path: %s", deployPath)
		return
	}
	
	deployDir := path.Dir(deployPath)
	deployName := path.Base(deployPath)
	//sourceName := path.Base(deploySource)
	serverName := fmt.Sprintf("%s(%s)", server.Name, server.Host)
	tarName := fmt.Sprintf("%s.tar.gz", deployName)
	workspace := vars[constant.WORKSPACE]
	sourceDir := path.Dir(deploySource)
	sourceName := path.Base(deploySource)
	log.Println(chalk.Green.Color("[upload target]"), deploySource)
	err = execer.RunCommand("", sourceDir, fmt.Sprintf("tar -czf %s %s", tarName, sourceName))
	//fmt.Sprintf("tar -czf %s %s && mv %s %s/%s", tarName, sourceName, tarName, workspace, tarName),
	if err != nil {
		return
	}
	
	info, err := sftpClient.Stat(deployDir)
	if err == nil {
		if !info.IsDir() {
			err = fmt.Errorf("server %s deploy path '%s' is not directory", serverName, deployDir)
			return
		}
	} else {
		if os.IsNotExist(err) {
			err = sftpClient.MkdirAll(deployDir)
			if err != nil {
				err = fmt.Errorf("server %s make path '%s' error: %s", serverName, deployDir, err)
				return
			}
		} else {
			logger.Error("GetTask path info error: ", err)
			return
		}
	}
	
	targetFile := filepath.Join(deployDir, tarName)
	file, err := sftpClient.Create(targetFile)
	if err != nil {
		logger.Error("Create bin file error: ", err)
		return
	}
	
	data, err := storage.Read(filepath.Join(workspace, tarName))
	if err != nil {
		logger.Error("Read bin file error: ", err)
		return
	}
	
	_, err = file.Write(data)
	if err != nil {
		logger.Error("Write bin file error: ", err)
		return
	}
	
	if strings.Contains(deployName, "/") {
		err = fmt.Errorf("illeagel deploy target name")
		return
	}
	
	err = execer.ServerRunCommand(sshClient, fmt.Sprintf("cd %s && rm -rf %s", deployDir, deployName))
	if err != nil {
		return
	}
	fmt.Println(deployDir, tarName)
	err = execer.ServerRunCommand(sshClient, fmt.Sprintf("cd %s && tar -xzf %s", deployDir, tarName))
	if err != nil {
		return
	}
	
	err = execer.ServerRunCommand(sshClient, fmt.Sprintf("rm %s", targetFile))
	if err != nil {
		return
	}
	
	return
}
