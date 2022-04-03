package core

import (
	"fmt"
	"github.com/koyeo/snowflake"
	"github.com/koyeo/yo/constant"
	"github.com/koyeo/yo/execer"
	"github.com/koyeo/yo/logger"
	"github.com/koyeo/yo/storage"
	"github.com/pkg/sftp"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
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
	Pipeline             []*Task
}

func (p *Task) vars() map[string]interface{} {
	return map[string]interface{}{}
}

func (p *Task) Run(c *cli.Context) (err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	if len(p.Pipeline) > 0 {

		log.Println(chalk.Green.Color("[run pipeline]"), p.Name)

		for _, v := range p.Pipeline {
			err = v.build(pwd, p.makeVars(v, pwd))
			if err != nil {
				return
			}
		}

		for _, v := range p.Pipeline {
			err = v.deploy(c, p.makeVars(v, pwd))
			if err != nil {
				return
			}
		}
	} else {
		return p.run(c)
	}

	// clean
	_ = storage.Remove(filepath.Join(pwd, constant.NEST_WORKSPACE))

	return
}

func (p *Task) run(c *cli.Context) (err error) {
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

	err = p.deploy(c, vars)
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

func (p *Task) deploy(c *cli.Context, vars map[string]string) (err error) {

	defer func() {
		if err != nil {
			err = nil
			os.Exit(1)
		}
	}()

	for _, v := range p.DeployServer {
		err = p.deployServer(c, v, vars)
		if err != nil {
			logger.Error("[deploy error]", err)
			return
		}
	}
	return
}

func (p *Task) deployServer(c *cli.Context, server *Server, vars map[string]string) (err error) {

	log.Println(chalk.Green.Color("[deploy start]"), fmt.Sprintf("%s (%s)", chalk.Yellow.Color(server.Name), server.Host))
	defer func() {
		if err != nil {
			return
		}
		log.Println(chalk.Green.Color("[deploy end]"), fmt.Sprintf("%s (%s)", chalk.Yellow.Color(server.Name), server.Host))
	}()

	var sshClient *ssh.Client

	proxyAddress := c.String("socks")
	if proxyAddress != "" {
		logger.Successf("[use socks] %s", proxyAddress)
		sshClient, err = NewProxySSHClient(proxyAddress, server)
	} else {
		sshClient, err = NewSSHClient(server)
	}
	if err != nil {
		return
	}
	defer func() {
		_ = sshClient.Close()
	}()

	sftpClient, err := NewSFTPClient(sshClient)
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
	//workspace := vars[constant.WORKSPACE]
	sourceDir := path.Dir(deploySource)
	sourceName := path.Base(deploySource)
	sourcePath := filepath.Join(sourceDir, tarName)
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

	distInfo, err := os.Stat(sourcePath)
	if err != nil {
		return
	}

	distFile, err := os.Open(sourcePath)
	if err != nil {
		return
	}
	defer func() {
		_ = distFile.Close()
	}()

	size := 2 * 1024 * 1024
	buf := make([]byte, 1024*1024)
	total := ByteSize(distInfo.Size())
	uploaded := int64(0)
	for {
		n, _ := distFile.Read(buf)
		if n == 0 {
			break
		}
		uploaded += int64(n)
		if n < size {
			_, err = file.Write(buf[0:n])
		} else {
			_, err = file.Write(buf)
		}
		if err != nil {
			logger.Error("Write bin file error: ", err)
			return
		}

		fmt.Printf("\rtotal: %s uploaded: %s", total, ByteSize(uploaded))
	}
	fmt.Printf("\n")

	if strings.Contains(deployName, "/") {
		err = fmt.Errorf("illeagel deploy target name")
		return
	}

	err = execer.ServerRunCommand(sshClient, fmt.Sprintf("cd %s && rm -rf %s", deployDir, deployName))
	if err != nil {
		return
	}
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
