package list

import (
	"fmt"
	"github.com/gozelle/_color"
	"github.com/koyeo/nest/common"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"os"

	"github.com/spf13/cobra"
)

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "显示配置项",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}()
	conf, err := protocol.Load(common.DefaultConfigFile)
	if err != nil {
		return
	}

	fmt.Printf("%s %s\n", title("version:"), conf.Version)

	if len(conf.Tasks) > 0 {
		fmt.Printf("%s \n", title("tasks:"))
		for key, task := range conf.Tasks {
			fmt.Printf("  %-25s %s\n", _color.CyanString(key), _color.WhiteString(task.Comment))
		}
	}
	if len(conf.Envs) > 0 {
		fmt.Printf("%s \n", title("envs:"))
		for key, value := range conf.Envs {
			fmt.Printf("  %-25s %s\n", _color.CyanString(key), value)
		}
	}

	if len(conf.Servers) > 0 {
		fmt.Printf("%s \n", title("servers:"))
		for key, server := range conf.Servers {
			fmt.Printf("  %-35s %s\n",
				_color.CyanString(fmt.Sprintf("%s(%s)", key, server.Host)), _color.WhiteString(server.Comment))
		}
	}
}

func title(s string) string {
	return _color.New(_color.FgHiGreen, _color.Bold).Sprint(s)
}
