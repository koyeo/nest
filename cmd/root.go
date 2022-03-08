package cmd

import (
	"github.com/koyeo/nest/cmd/git"
	initCmd "github.com/koyeo/nest/cmd/init"
	"github.com/koyeo/nest/cmd/run"
	"github.com/koyeo/nest/cmd/upload"
	"github.com/koyeo/nest/cmd/watch"
	"os"
	
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nest",
	Short: "开发辅助命令行工具集",
	Long: `开发辅助命令行工具集，支持本地构建、服务器（代理）上传、远程命令执行、流水线任务等`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(
		initCmd.Cmd,
		run.Cmd,
		upload.Cmd,
		git.Cmd,
		watch.Cmd,
	)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nest.yaml)")
	
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
