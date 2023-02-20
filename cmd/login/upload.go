package upload

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"os"
)

var (
	uploadSrc  string
	uploadDist string
	uploadHost string
	uploadPort uint64
	uploadUser string
	uploadPem  string
	uploadYes  bool
)

// Cmd represents the upload command
var Cmd = &cobra.Command{
	Use:   "upload",
	Short: "上传文件或目录到远程服务器",
	Long:  `用于上传文件或目录到远程服务器,支持网络代理（http,socks5），远程目录询问自动创建`,
	Run: func(cmd *cobra.Command, args []string) {
		err := login(cmd)
		if err != nil {
			if err.Error() != "exit status 1" {
				fmt.Println(chalk.Red.Color(err.Error()))
			}
			os.Exit(1)
			return
		}
	},
}

func init() {
	Cmd.PersistentFlags().StringVar(&uploadSrc, "src", "", "指定源文件或目录路径")
	Cmd.PersistentFlags().StringVar(&uploadDist, "dist", "", "指定远程存放目录路径")
	Cmd.PersistentFlags().StringVar(&uploadHost, "host", "", "远程服务器 IP 地址，如：192.168.1.2")
	Cmd.PersistentFlags().StringVar(&uploadUser, "user", "", "远程服务器用户，如：root")
	Cmd.PersistentFlags().Uint64Var(&uploadPort, "port", 22, "远程服务器端口，默认采用 sftp 上传使用 22 端口")
	Cmd.PersistentFlags().StringVar(&uploadPem, "pem", "~/.ssh/id_rsa", "服务器认证私钥文件，默认使用: ~/.ssh/id_rsa 文件")
	//Cmd.PersistentFlags().BoolVar(&uploadYes, "y", false, "开启默允模式，如果远程目录不存在，则自动创建，若远程目录下同名文件和目录已存在，则自动覆盖")
}

// 执行上传
func login(cmd *cobra.Command) error {
	
	return nil
}
