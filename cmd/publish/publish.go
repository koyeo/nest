package publish

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"io"
	"net"
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
	Use:   "publish",
	Short: "上传文件或目录到远程服务器",
	Long:  `用于上传文件或目录到远程服务器,支持网络代理（http,socks5），远程目录询问自动创建`,
	Run: func(cmd *cobra.Command, args []string) {
		err := publish(cmd)
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
	//Cmd.PersistentFlags().StringVar(&uploadSrc, "src", "", "指定源文件或目录路径")
	//Cmd.PersistentFlags().StringVar(&uploadDist, "dist", "", "指定远程存放目录路径")
	//Cmd.PersistentFlags().StringVar(&uploadHost, "host", "", "远程服务器 IP 地址，如：192.168.1.2")
	//Cmd.PersistentFlags().StringVar(&uploadUser, "user", "", "远程服务器用户，如：root")
	//Cmd.PersistentFlags().Uint64Var(&uploadPort, "port", 22, "远程服务器端口，默认采用 sftp 上传使用 22 端口")
	//Cmd.PersistentFlags().StringVar(&uploadPem, "pem", "~/.ssh/id_rsa", "服务器认证私钥文件，默认使用: ~/.ssh/id_rsa 文件")
	//Cmd.PersistentFlags().BoolVar(&uploadYes, "y", false, "开启默允模式，如果远程目录不存在，则自动创建，若远程目录下同名文件和目录已存在，则自动覆盖")
}

// 执行上传
func publish(cmd *cobra.Command) (err error) {
	file, err := os.Open("nest.json")
	if err != nil {
		
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	
	conn, err := net.Dial("tcp", "127.0.0.1:8332")
	if err != nil {
		fmt.Println("net.Dialt err", err)
		return
	}
	//发送文件名到接收端
	_, err = conn.Write([]byte("nest.json"))
	if err != nil {
		fmt.Println("conn.Write err", err)
		return
	}
	buf := make([]byte, 4096)
	//接收服务器返还的指令
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("conn.Read err", err)
		return
	}
	//返回ok，可以传输文件
	if string(buf[:n]) == "ok" {
		sendFile(conn, "nest.json")
	}
	return
}

func sendFile(conn net.Conn, filepath string) {
	//打开要传输的文件
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("os.Open err", err)
		return
	}
	buf := make([]byte, 4096)
	//循环读取文件内容，写入远程连接
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			fmt.Println("文件读取完毕")
			fmt.Println("文件传输完毕")
			return
		}
		if err != nil {
			fmt.Println("file.Read err:", err)
			return
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Println("conn.Write err:", err)
			return
		}
		fmt.Println("写入文件")
	}
}
