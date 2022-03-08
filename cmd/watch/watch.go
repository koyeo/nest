package watch

import (
	"fmt"
	
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "watch",
	Short: "监听文件或目录变动触发命令行执行",
	Long:  `监听文件或目录变动触发命令行执行，在 nest.toml 任务配置模式下，可以根据文件变化触发任务执行`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("watch called")
	},
}
