package upload

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	src  string
	dist string
	host string
	port int64
	yes  bool
)

// Cmd represents the upload command
var Cmd = &cobra.Command{
	Use:   "upload",
	Short: "上传文件或目录到远程服务器，支持网络代理（http,socks5），远程目录询问自动创建",
	Long:  `用于上传文件或目录到远程服务器`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return upload(cmd)
	},
}

func init() {
	Cmd.PersistentFlags().StringVar(&src, "src", "", "指定源文件或目录路径")
	Cmd.PersistentFlags().StringVar(&dist, "dist", "", "指定远程目录路径")
	Cmd.PersistentFlags().StringVar(&host, "host", "", "远程服务器 IP 地址，如：192.168.1.2")
	Cmd.PersistentFlags().Int64Var(&port, "port", 22, "远程服务器端口，默认采用 sftp 上传使用 22 端口")
	Cmd.PersistentFlags().BoolVar(&yes, "y", false, "开启默允模式，如果远程目录不存在，则自动创建，若盖目录下同名文件和目录已存在，则自动覆盖")
}

func upload(cmd *cobra.Command) error {
	if src == "" {
		return fmt.Errorf("请指定待上传的源文件或目录路径，如 --src=./demo.txt")
	}
	if dist == "" {
		return fmt.Errorf("请指定远程存放目录路径，如 --dist=/app/demo")
	}
	if host == "" {
		return fmt.Errorf("请指定远程服务器地址，如 --host=192.168.1.168")
	}
	return nil
}
