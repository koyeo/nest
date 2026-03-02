package initialize

import (
	"fmt"
	"strings"

	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/common"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project / 项目初始化",
	Long:  `Initialize nest.yaml and inject .gitignore config. / 初始化 nest.yaml 文件，注入 .gitignore 配置。`,
	RunE:  initialize,
}

func initialize(cmd *cobra.Command, args []string) (err error) {

	configFile := common.DefaultConfigFile
	l := len(args)
	if l > 1 {
		err = fmt.Errorf("at most accept on file")
		return
	} else if l == 1 {
		configFile = args[0]
	}

	ok, err := _fs.Exists(configFile)
	if err != nil {
		return
	}
	if !ok {
		err = _fs.Write(configFile, []byte(strings.TrimSpace(tpl)))
		if err != nil {
			return
		}
		fmt.Printf("create %s\n", configFile)
	} else {
		fmt.Printf("%s already exists\n", configFile)
	}

	err = injectGitIgnore()
	if err != nil {
		return
	}

	return
}

const tpl = `
##########################################################
#                        Nest                            #
#            用于快速部署本地本地构建部署工具                 #
##########################################################
version: 1.0
servers:
  server-1:
    comment: 第一个服务器
    host: 192.168.1.10 
    user: root         
    # 默认使用 ~/.ssh/id_rsa 私钥进行认证
    # identity_file: ~/.ssh/id_rsa
    # password: 123456
# 声明此文件用到的云存储，key 为本文件中引用的别名，value 为全局配置名称
# 全局配置通过 nest storage add 命令添加
storage:
  oss: oss
tasks:
  task-1:
    comment: 第一个任务
    steps:
      - run: echo "Hi! this is from first task!"
  # 任务名称 nest run 执行标识
  task-2:                                      
    # 任务注释
    comment: 第二个任务                       
    steps:
      # 引用其它任务的全部 steps
      - use: task-1
      # 在本地执行命令
      - run: echo 'Hi! this is from second task!'            
      # 部署服务器
      - deploy:
          servers:
            # 引用服务器
            - use: server-1                    
          files:                             
            # 本地文件路径
            - source: ./hi.txt                    
            # 服务器存放位置		
              target: /app/foo/hi         
          executes:
            # 在服务器执行命令
            - run: echo 'Hi! this is from server-1'
      # 在本地执行命令
      - run: echo 'Hi! this is from local!'
`

func injectGitIgnore() (err error) {

	const gitignore = ".gitignore"
	ok, err := _fs.Exists(gitignore)
	if err != nil {
		return
	}
	if !ok {
		err = _fs.Write(gitignore, []byte(common.TmpWorkspace))
		if err != nil {
			return
		}
		fmt.Printf("create %s\n", gitignore)
		return
	}

	content, err := _fs.Read(gitignore)
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	exists := false
	for _, line := range lines {
		if strings.TrimSpace(line) == common.TmpWorkspace {
			exists = true
			break
		}
	}
	if !exists {
		lines = append(lines, common.TmpWorkspace)
		err = _fs.Write(gitignore, []byte(strings.Join(lines, "\n")))
		if err != nil {
			return
		}
		fmt.Printf("update %s\n", gitignore)
		return
	}

	return
}
