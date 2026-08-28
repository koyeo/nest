import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { Command } from "commander";
import { DEFAULT_CONFIG_FILE, TMP_WORKSPACE } from "../common/const.js";

const TEMPLATE = `# ─────────────────────────────────────────────
#  Ship — Task Runner & Deployment Config
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
# Cloud object storage references. Key = alias used in this file.
# Value is either a global config name (added via "ship storage add", stays on this machine):
# storages:
#   oss: my-oss-config
# ...or inline credentials so the file can be shared (generate with "ship storage encrypt").
# The encryption key is built into ship: anyone holding this file can access the bucket,
# so treat the bucket as public and clean it regularly ("ship storage clean <alias>").
# storages:
#   oss:
#     provider: oss
#     endpoint: oss-cn-hangzhou.aliyuncs.com
#     bucket: my-bucket
#     access_key_id: enc:...
#     access_key_secret: enc:...

# ── Environment Variables ──
# Global env vars available to all tasks. Task-level envs override these.
# envs:
#   NODE_ENV: production

# ── Tasks ──
# Named step sequences executed by "ship run <task-name>".
tasks:
  hello:
    comment: A simple demo task
    commands:
      - run: echo "Hello from ship!"

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
          # Set "storage: <alias>" to transfer via cloud storage instead
          # (falls back to SFTP automatically if the storage is unavailable).
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
`;

export function initCommand(): Command {
  return new Command("init")
    .description(`Initialize a ${DEFAULT_CONFIG_FILE} config file and update .gitignore`)
    .argument("[config-file]", "config file name", DEFAULT_CONFIG_FILE)
    .action((configFile: string) => {
      if (existsSync(configFile)) {
        process.stdout.write(`${configFile} already exists\n`);
      } else {
        writeFileSync(configFile, TEMPLATE);
        process.stdout.write(`create ${configFile}\n`);
      }
      injectGitIgnore();
    });
}

function injectGitIgnore(): void {
  const gitignore = ".gitignore";
  if (!existsSync(gitignore)) {
    writeFileSync(gitignore, TMP_WORKSPACE);
    process.stdout.write(`create ${gitignore}\n`);
    return;
  }
  const lines = readFileSync(gitignore, "utf8").split("\n");
  if (lines.some((line) => line.trim() === TMP_WORKSPACE)) {
    return;
  }
  lines.push(TMP_WORKSPACE);
  writeFileSync(gitignore, lines.join("\n"));
  process.stdout.write(`update ${gitignore}\n`);
}
