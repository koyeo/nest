package initialize

import (
	"fmt"
	"strings"

	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/common"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "init [config-file]",
	Short: "Initialize a nest.yaml config file and update .gitignore",
	Long: `Create a starter nest.yaml config file in the current directory.
If the file already exists, the command is a no-op.

Also injects ".nest" into .gitignore (creates the file if needed).

The generated template includes annotated examples of all supported features:
servers, tasks, multi-line run, use (task composition), upload, deploy
(with cwd, shell_init, conflict_strategy), and storage references.

Examples:
  nest init                    # Create nest.yaml
  nest init nest.prod.yaml     # Create a custom-named config`,
	RunE: initialize,
}

func initialize(cmd *cobra.Command, args []string) (err error) {

	configFile := common.DefaultConfigFile
	l := len(args)
	if l > 1 {
		err = fmt.Errorf("at most accept on file")
		return
	} else if l == 1 {
		configFile = args[0]
	}

	ok, err := _fs.Exists(configFile)
	if err != nil {
		return
	}
	if !ok {
		err = _fs.Write(configFile, []byte(strings.TrimSpace(tpl)))
		if err != nil {
			return
		}
		fmt.Printf("create %s\n", configFile)
	} else {
		fmt.Printf("%s already exists\n", configFile)
	}

	err = injectGitIgnore()
	if err != nil {
		return
	}

	return
}

const tpl = `
# ─────────────────────────────────────────────
#  Nest — Task Runner & Deployment Config
# ─────────────────────────────────────────────
version: 1.0

# ── Servers ──
# Named SSH connection profiles used in deploy commands.
# Auth: defaults to ~/.ssh/id_rsa if neither password nor identity_file is set.
servers:
  my-server:
    host: 192.168.1.10
    user: root
    # port: 22
    # identity_file: ~/.ssh/id_rsa
    # password: secret

# ── Storage ──
# Cloud object storage references. Key = alias used in this file,
# value = global config name (added via "nest storage add").
# storages:
#   oss: my-oss-config

# ── Environment Variables ──
# Global env vars available to all tasks. Task-level envs override these.
# envs:
#   NODE_ENV: production

# ── Tasks ──
# Named step sequences executed by "nest run <task-name>".
tasks:
  hello:
    comment: A simple demo task
    commands:
      - run: echo "Hello from nest!"

  build:
    comment: Example multi-line local build
    commands:
      # Multi-line commands are supported via YAML literal blocks (|).
      # Each line runs sequentially in the same shell session.
      - run: |
          echo "Building project..."
          mkdir -p dist
          echo "Build complete"

  deploy:
    comment: Full deploy example (build + upload + remote setup)
    commands:
      # Reuse commands from another task
      - use: build

      # Upload local artifacts to cloud storage (requires storage config)
      # - upload:
      #     storage: oss
      #     source: ./dist

      # Deploy to remote servers
      - deploy:
          servers:
            - use: my-server

          # Transfer files to the server.
          # Default: direct SFTP upload (tar + extract).
          # Set "storage: <alias>" to transfer via cloud storage instead.
          files:
            - source: ./dist
              target: /data/app
              # storage: oss   # Use cloud storage alias "oss" for transfer

          # Working directory for all execute commands (optional)
          # cwd: /data/app

          # Shell init command prepended to each execute (optional)
          # Useful for loading nvm, pyenv, etc.
          # shell_init: source /root/.nvm/nvm.sh


          # Commands to run on each server after file upload
          commands:
            - run: echo "Deployed successfully to $(hostname)"
`

func injectGitIgnore() (err error) {

	const gitignore = ".gitignore"
	ok, err := _fs.Exists(gitignore)
	if err != nil {
		return
	}
	if !ok {
		err = _fs.Write(gitignore, []byte(common.TmpWorkspace))
		if err != nil {
			return
		}
		fmt.Printf("create %s\n", gitignore)
		return
	}

	content, err := _fs.Read(gitignore)
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	exists := false
	for _, line := range lines {
		if strings.TrimSpace(line) == common.TmpWorkspace {
			exists = true
			break
		}
	}
	if !exists {
		lines = append(lines, common.TmpWorkspace)
		err = _fs.Write(gitignore, []byte(strings.Join(lines, "\n")))
		if err != nil {
			return
		}
		fmt.Printf("update %s\n", gitignore)
		return
	}

	return
}
