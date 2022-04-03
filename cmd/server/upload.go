package server

import (
	"fmt"
	"github.com/koyeo/yo/execer"
	"github.com/koyeo/yo/logger"
	"github.com/koyeo/yo/storage"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"strings"
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
	Use:   "server",
	Short: "上传文件或目录到远程服务器，支持网络代理（http,socks5），远程目录询问自动创建",
	Long:  `用于上传文件或目录到远程服务器`,
	Run: func(cmd *cobra.Command, args []string) {
		err := upload(cmd)
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
	Cmd.PersistentFlags().BoolVar(&uploadYes, "y", false, "开启默允模式，如果远程目录不存在，则自动创建，若远程目录下同名文件和目录已存在，则自动覆盖")
}

// 执行上传
func upload(cmd *cobra.Command) error {
	if uploadSrc == "" {
		return fmt.Errorf("请指定待上传的源文件或目录路径，如 --src=./demo.txt")
	}
	if uploadDist == "" {
		return fmt.Errorf("请指定远程存放目录路径，如 --dist=/app/demo")
	}
	if uploadHost == "" {
		return fmt.Errorf("请指定远程服务器地址，如 --host=192.168.1.168")
	}
	if uploadUser == "" {
		return fmt.Errorf("请指定远程服务器用户，如 --user=root")
	}
	server := PrepareUploadServer()
	err := Upload(server)
	if err != nil {
		return err
	}
	return nil
}

type Server struct {
	Name         string
	Host         string
	Port         uint64
	User         string
	Password     string
	IdentityFile string
}

func Upload(server *Server) (err error) {
	logger.Print(fmt.Sprintf(
		"%s %s(%s)\n",
		chalk.Green.Color("[上传开始]"),
		chalk.Yellow.Color(server.Name),
		server.Host),
	)
	defer func() {
		if err != nil {
			//logger.Print(fmt.Sprintf("%s %s\n",
			//	chalk.Red.Color("[上传错误]"),
			//	err,
			//))
			return
		}
		logger.Print(fmt.Sprintf("%s %s(%s)\n",
			chalk.Green.Color("[上传完成]"),
			chalk.Yellow.Color(server.Name),
			server.Host,
		))
	}()

	// 构造 Sftp 对象用以服务器操作
	sshClient, sftpClient, err := PrepareSftpClient(server)
	if err != nil {
		return
	}
	defer func() {
		_ = sftpClient.Close()
		_ = sshClient.Close()
	}()

	// 测试服务器连通性
	logger.Print(chalk.Cyan.Color("[上传中] 测试服务器连通性... "))
	err = ping(sshClient)
	if err != nil {
		return
	}
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	logger.Print(chalk.Cyan.Color("[上传中] 检查远程目录是否存在... "))
	err = checkDistDir(sshClient, uploadDist)
	if err != nil {
		return
	}
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	logger.Print(chalk.Cyan.Color("[上传中] 检查远程目录下同名文件或目录是否存在... "))
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	logger.Print(chalk.Cyan.Color("[上传中] 压缩本地源文件... "))
	err = compressSrc(uploadSrc)
	if err != nil {
		return
	}
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	logger.Print(chalk.Cyan.Color("[上传中] 开始上传... "))
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	logger.Print(chalk.Cyan.Color("[上传中] 解压远程文件... "))
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	logger.Print(chalk.Cyan.Color("[上传中] 清理临时文件... "))
	fmt.Printf(chalk.Cyan.Color("ok\n"))

	// 打印上传信息

	return
}

// PrepareUploadServer  准备 Server 对象
func PrepareUploadServer() (server *Server) {
	server = &Server{
		Host:         uploadHost,
		Port:         uploadPort,
		User:         uploadUser,
		IdentityFile: uploadPem,
	}
	return
}

// PrepareSftpClient 准备 ssh 和 sftp 客户端
// 使用完毕后请注意调用： sshClient.Close()、sftpClient.Close()
func PrepareSftpClient(server *Server) (sshClient *ssh.Client, sftpClient *sftp.Client, err error) {
	// 构造 ssh Client
	// TODO 处理网络代理
	sshClient, err = NewSSHClient(server)
	if err != nil {
		err = fmt.Errorf("服务器连接错误: %s", err)
		return
	}
	// 构造 sftp Client
	sftpClient, err = NewSFTPClient(sshClient)
	if err != nil {
		_ = sshClient.Close()
		return
	}
	return
}

// 测试 ssh 客户端连通性
func ping(sshClient *ssh.Client) (err error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return
	}
	result, err := session.Output(`echo "ok"`)
	if err != nil {
		return
	}
	ok := strings.TrimSpace(string(result))
	if ok != "ok" {
		err = fmt.Errorf("expect ping result: ok, got: %s", ok)
		return
	}
	return
}

//  执行远程命令
func exec(sshClient *ssh.Client) {

}

func checkDistDir(sshClient *ssh.Client, dist string) (err error) {

	return
}

// 压缩本地源文件
func compressSrc(src string) (err error) {
	target := fmt.Sprintf("%s/%s.tar.gz", PrepareGetNestTempDir(), path.Base(src))
	err = execer.RunCommand("", "", fmt.Sprintf("tar -czf %s %s", target, src))
	if err != nil {
		return
	}
	return
}

// 解压远程目标文件
func decompressDist() {
	fmt.Println(uploadDist)
}
func GetNestTempDir() string {
	return "./.nest/temp"
}
func PrepareGetNestTempDir() string {
	dir := GetNestTempDir()
	if !storage.Exist(dir) {
		storage.MakeDir(dir)
	}
	return dir
}

func CleanNestTempDir() {
	_ = storage.Remove(GetNestTempDir())
}
