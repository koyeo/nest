package initialize

import (
	"fmt"
	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/common"
	"github.com/spf13/cobra"
	"strings"
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "项目初始化",
	Long:  `初始化 nest.yml 文件，注入 .gitignore 配置`,
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
#              https://nest.kozilla.io                   #
##########################################################
version: 1.0
servers:
  server-1:
    comment: 示例服务器
    host: 192.168.1.10
    user: root                                 # 默认使用 ~/.ssh/id_rsa 私钥进行认证
tasks:
  task-1:                                      # 任务名称
    comment: 示例任务                           # 任务注释
    steps:
      - run: go build -o foo foo.go            # 本地执行构建
      - deploy:
          servers:
            - use: server-1                    # 部署服务器
          mappers:                             
            - source: ./foo                    # 本地文件路径
              target: /app/foo/bin/foo         # 服务器存放位置
          executes:
            - run: supervisorctl restart foo   # 服务器重启服务
      - run: rm foo                            # 本地清理
  hi:
    comment: 打个招呼
    steps:
      - run: echo "Hi! this is from nest~"
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
