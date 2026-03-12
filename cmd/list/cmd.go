package list

import (
	"fmt"
	"os"

	"github.com/gozelle/_color"
	"github.com/koyeo/nest/common"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"

	"github.com/spf13/cobra"
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "Display tasks, servers, and environment variables from the config",
	Long: `Parse the nest config file and display a summary of all declared
tasks (with comments), servers (with hosts), and environment variables.

Useful for discovering available task names before running them.

Example:
  nest list
  nest list -c nest.prod.yaml`,
	Run: run,
}

func run(cmd *cobra.Command, args []string) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(err)
			os.Exit(1)
		}
	}()
	conf, err := protocol.Load(common.ConfigFile)
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
