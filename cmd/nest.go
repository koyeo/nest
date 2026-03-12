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
	Short: "Task runner & deployment CLI for local builds, remote execution, and server deploys",
	Long: `Nest is a YAML-driven task runner and deployment tool.

It reads a config file (default: nest.yaml) that declares servers, tasks, and
cloud storage, then executes multi-step pipelines that can mix local commands,
remote SSH execution, file uploads (SCP or cloud storage), and deployment with
conflict resolution.

Architecture:
  nest.yaml ──► TaskRunner ──► local exec (bash)
                             ──► SSH remote exec (login shell)
                             ──► file deploy (tar + SCP or OSS presigned URL)

Key Concepts:
  servers     Named SSH connection profiles (host, user, auth).
  tasks       Named step sequences. Steps can be:
                • run:      Execute a shell command (local or remote).
                            Supports multi-line YAML literal blocks (|).
                • use:      Include all steps from another task (composition).
                • upload:   Compress & upload a local path to cloud storage.
                • deploy:   Upload files to remote servers and run commands.
  deploy      A deploy step contains:
                • servers:            Target servers (by reference).
                • files:              File mappings (source → target).
                                      Default transfer: SFTP (tar + extract).
                                      Set "storage: <alias>" on a file entry to
                                      transfer via cloud storage instead.
                • executes:           Commands to run on each server after upload.
                • cwd:                Working directory for all execute commands.
                • shell_init:         Shell init command (e.g. "source ~/.nvm/nvm.sh")
                                      prepended to each execute command.
                • conflict_strategy:  How to handle file conflicts on the server:
                                      "overwrite", "backup", or "error".
                                      If unset, prompts interactively.
  storage     Cloud object storage configs (OSS / S3). Credentials are encrypted
              and stored globally in ~/.nest/config.json. Referenced in nest.yaml
              via aliases declared in the "storage:" section.

Config file (nest.yaml) schema:
  version: 1.0
  servers:   { name: { host, port, user, password, identity_file } }
  storage:   { alias: global_config_name }
  envs:      { KEY: VALUE }
  tasks:     { name: { comment, workspace, envs, steps: [...] } }

Flags:
  -c, --config   Path to config file (default: nest.yaml)

Use "nest [command] --help" for more information about a command.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version, commit hash, and build time",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nest %s\n", appVersion)
		fmt.Printf("  commit:  %s\n", appCommit)
		fmt.Printf("  built:   %s\n", appBuild)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&common.ConfigFile, "config", "c", common.DefaultConfigFile,
		"Path to nest config file (default: nest.yaml)")
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
