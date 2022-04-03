package net

import (
	"github.com/spf13/cobra"
)

// NetCmd  represents the init command
var NetCmd = &cobra.Command{
	Use:   "net",
	Short: "网络配置命令",
	Long:  `在项目根目录下创建 nest.toml 文件，用于配置任务信息`,
	RunE: func(cmd *cobra.Command, args []string) error {
		//return initConfig(cmd)
		return nil
	},
}
