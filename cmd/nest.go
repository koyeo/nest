package cmd

import (
	"github.com/koyeo/nest/cmd/initialize"
	"github.com/koyeo/nest/cmd/list"
	"github.com/koyeo/nest/cmd/publish"
	"github.com/koyeo/nest/cmd/run"
	"github.com/koyeo/nest/cmd/upload"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nest",
	Short: "开发辅助命令行工具集",
	Long:  `开发辅助命令行工具集，支持本地构建、服务器（代理）上传、远程命令执行、流水线任务等`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(
		initialize.Cmd,
		run.Cmd,
		list.Cmd,
		upload.Cmd,
		publish.Cmd,
	)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
