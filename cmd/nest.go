package cmd

import (
	"os"

	"github.com/koyeo/nest/cmd/bucket"
	"github.com/koyeo/nest/cmd/initialize"
	"github.com/koyeo/nest/cmd/list"
	"github.com/koyeo/nest/cmd/run"
	"github.com/koyeo/nest/common"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nest",
	Short: "Development helper CLI / 开发辅助命令行工具集",
	Long: `Nest - Development helper CLI for local build, server upload, remote exec & pipeline tasks.
Nest - 开发辅助命令行工具集，支持本地构建、服务器上传、远程命令执行、流水线任务等。`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&common.ConfigFile, "config", "c", common.DefaultConfigFile,
		"Config file path / 指定配置文件路径")
}

func Execute() {
	rootCmd.AddCommand(
		initialize.Cmd,
		run.Cmd,
		list.Cmd,
		bucket.Cmd,
		//upload.Cmd,
	)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
