package cmd

import (
	"fmt"
	"os"

	"github.com/koyeo/nest/cmd/initialize"
	"github.com/koyeo/nest/cmd/list"
	"github.com/koyeo/nest/cmd/run"
	"github.com/koyeo/nest/cmd/storagecmd"
	"github.com/koyeo/nest/common"
	"github.com/spf13/cobra"
)

var (
	appVersion = "dev"
	appCommit  = "unknown"
	appBuild   = "unknown"
)

// SetVersionInfo is called from main to inject ldflags values.
func SetVersionInfo(version, commit, buildTime string) {
	appVersion = version
	appCommit = commit
	appBuild = buildTime
}

var rootCmd = &cobra.Command{
	Use:   "nest",
	Short: "Development helper CLI / 开发辅助命令行工具集",
	Long: `Nest - Development helper CLI for local build, server upload, remote exec & pipeline tasks.
Nest - 开发辅助命令行工具集，支持本地构建、服务器上传、远程命令执行、流水线任务等。`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information / 查看版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nest %s\n", appVersion)
		fmt.Printf("  commit:  %s\n", appCommit)
		fmt.Printf("  built:   %s\n", appBuild)
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
		storagecmd.Cmd,
		versionCmd,
		//upload.Cmd,
	)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
