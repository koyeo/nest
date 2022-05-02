package initialize

import (
	"github.com/spf13/cobra"
)

// Cmd  represents the init command
var Cmd = &cobra.Command{
	Use:   "init",
	Short: "项目初始化",
	Long:  `在项目根目录下创建 nest.toml 文件，用于配置任务信息`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},
}

func initConfig(cmd *cobra.Command) error {
	//path := c.String(constant.NEST_TOML)
	//if path != "" {
	//	return config.Load(path)
	//}
	//// TODO 指定项目
	//// TODO 指定配置文件
	//pwd, err := os.Getwd()
	//if err != nil {
	//	return err
	//}
	//err = config.Load(filepath.Join(pwd, constant.NEST_TOML))
	//if err != nil {
	//	return err
	//}
	//
	//err = core.Project.LoadConfig()
	//if err != nil {
	//	return err
	//}
	return nil
}
