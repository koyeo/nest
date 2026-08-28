package main

import "github.com/koyeo/nest/cmd"

// Set via -ldflags at build time
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, buildTime)
	cmd.Execute()
}
