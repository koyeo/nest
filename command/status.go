package command

import (
	"github.com/urfave/cli"
	"nest/core"
)

func StatusCommand(c *cli.Context) error {
	core.Status()
	return nil
}
