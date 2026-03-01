<p align="center">
  <img src="./logo.png" alt="Nest" width="200" />
</p>

# Nest

[中文文档](./README_zh.md)

A lightweight local CI/CD tool for rapid integration and deployment. For small to mid-sized projects, Nest can fully replace traditional DevOps tools — handling builds, deployments, server management, and operational tasks all from your local machine.

## Why Nest?

There are many CI/CD tools out there — Jenkins, GitHub Actions, Makefile, Ansible, Travis CI, etc. But they're often too heavy for quick delivery:

- **Jenkins** requires a full server deployment and web UI.
- **GitHub Actions / Travis CI** are platform-specific and require pushing code to trigger.
- **Ansible + Makefile** need to be configured together every time.

Nest solves these pain points with a **single YAML config** and a **single CLI command**.

### Best For

- Solo or small-team full-stack developers who want fast feedback loops.
- Quick deployments of frontend/backend projects to servers.
- Lightweight server ops: log checks, service restarts, DB backups, cert renewals.
- Multi-environment management (dev / staging / production) with separate config files.

### Not Ideal For

- Large-scale production environments requiring strict release management and approval workflows.

## Installation

### Quick Install (Recommended)

Download the latest binary for your platform:

```bash
# macOS (Apple Silicon)
curl -fsSL https://github.com/koyeo/nest/releases/latest/download/nest-darwin-arm64 -o /usr/local/bin/nest && chmod +x /usr/local/bin/nest

# macOS (Intel)
curl -fsSL https://github.com/koyeo/nest/releases/latest/download/nest-darwin-amd64 -o /usr/local/bin/nest && chmod +x /usr/local/bin/nest

# Linux (x86_64)
curl -fsSL https://github.com/koyeo/nest/releases/latest/download/nest-linux-amd64 -o /usr/local/bin/nest && chmod +x /usr/local/bin/nest

# Linux (ARM64)
curl -fsSL https://github.com/koyeo/nest/releases/latest/download/nest-linux-arm64 -o /usr/local/bin/nest && chmod +x /usr/local/bin/nest
```

To **update** to the latest version, run the same command again.

### Install via Go

```bash
go install github.com/koyeo/nest@latest
```

> **Note:** Make sure `$GOPATH/bin` is in your `$PATH`.

## Quick Start

### 1. Initialize Configuration

```bash
nest init
```

This creates a `nest.yaml` and adds `.nest` to `.gitignore`.

### 2. Edit `nest.yaml`

Here's a practical example — build a Go backend, deploy to server, and restart the service:

```yaml
version: 1.0

servers:
  prod:
    comment: Production server
    host: 192.168.1.10
    user: root
    # identity_file: ~/.ssh/id_rsa    # Default SSH key
  staging:
    comment: Staging server
    host: 192.168.1.20
    user: deploy
    port: 2222

envs:
  APP_NAME: myapp
  REMOTE_DIR: /opt/myapp

tasks:
  # ── Build & Deploy ────────────────────────────────────
  deploy:
    comment: Build and deploy to production
    steps:
      - run: echo "🔨 Building..."
      - run: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp .
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
            - source: ./configs/prod.yaml
              target: /opt/myapp/configs/app.yaml
          executes:
            - run: chmod +x /opt/myapp/bin/myapp
            - run: systemctl restart myapp
      - run: rm -f myapp
      - run: echo "✅ Deployed!"

  # ── Frontend Deploy ───────────────────────────────────
  deploy-web:
    comment: Build frontend and deploy static files
    steps:
      - run: cd frontend && npm ci && npm run build
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./frontend/dist/
              target: /var/www/myapp/
          executes:
            - run: nginx -s reload

  # ── Server Operations ────────────────────────────────
  logs:
    comment: Tail production logs
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: journalctl -u myapp -f --lines=100

  status:
    comment: Check service status on all servers
    steps:
      - deploy:
          servers:
            - use: prod
            - use: staging
          executes:
            - run: systemctl status myapp && df -h && free -m

  restart:
    comment: Restart service
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: systemctl restart myapp
            - run: echo "✅ Service restarted"

  # ── Database ──────────────────────────────────────────
  db-backup:
    comment: Backup production database
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: |
                TIMESTAMP=$(date +%Y%m%d_%H%M%S)
                pg_dump -U postgres myapp_db > /opt/backups/myapp_${TIMESTAMP}.sql
                echo "✅ Backup saved: myapp_${TIMESTAMP}.sql"

  db-migrate:
    comment: Run database migrations
    steps:
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./migrations/
              target: /opt/myapp/migrations/
          executes:
            - run: cd /opt/myapp && ./bin/myapp migrate up

  # ── SSL Certificate ───────────────────────────────────
  cert-renew:
    comment: Renew SSL certificates
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: certbot renew --quiet && nginx -s reload

  # ── Full Pipeline ─────────────────────────────────────
  release:
    comment: Full release pipeline — test, build, deploy, verify
    steps:
      - run: go test ./...
      - use: deploy
      - deploy:
          servers:
            - use: prod
          executes:
            - run: curl -sf http://localhost:8080/health || exit 1
            - run: echo "✅ Health check passed"
```

### 3. Run Tasks

```bash
nest run deploy            # Build & deploy
nest run logs              # Tail server logs
nest run status            # Check server status
nest run db-backup         # Backup database
nest run release           # Full release pipeline
nest run deploy deploy-web # Run multiple tasks
```

## Multi-Environment with `--config`

Use the `-c` / `--config` flag to manage different environments with separate config files:

```bash
nest init                          # Creates nest.yaml (default)
nest init nest.staging.yml         # Create staging config
nest init nest.production.yml      # Create production config
```

```bash
nest run deploy                    # Uses nest.yaml (default / dev)
nest run deploy -c nest.staging.yml       # Deploy to staging
nest run deploy -c nest.production.yml    # Deploy to production
nest list -c nest.production.yml          # List production config
```

This makes it easy to maintain isolated configs per environment while sharing the same task definitions.

## CLI Reference

| Command | Description |
|:--------|:------------|
| `nest init [file]` | Initialize config file and update `.gitignore` |
| `nest run <task...>` | Execute one or more tasks by name |
| `nest list` | List all configured resources |

### Global Flags

| Flag | Short | Description |
|:-----|:------|:------------|
| `--config <file>` | `-c` | Specify config file (default: `nest.yaml`) |

## Configuration Reference

### Servers

```yaml
servers:
  my_server:
    comment: My server                # Description
    host: 192.168.1.5                 # Server address
    port: 2222                        # Port (default: 22)
    user: root                        # Username
    password: 123456                  # Password auth
    identity_file: ~/.ssh/id_rsa      # Key auth (default)
```

### Environment Variables

```yaml
envs:
  APP_NAME: myapp
  VERSION: "1.0.0"
```

### Deploy File Mapping

| source  | target            | Remote result             |
|:--------|:------------------|:--------------------------|
| `file1` | `/app/test/file1` | `/app/test/file1`         |
| `file1` | `/app/test/file2` | `/app/test/file2`         |
| `file1` | `/app/test`       | `/app/test`               |
| `file1` | `/app/test/`      | `/app/test/file1`         |
| `dir1`  | `/app/test/dir2`  | `/app/test/dir2`          |
| `dir1`  | `/app/test/dir2/` | `/app/test/dir1/dir2`     |
| `dir1`  | `/app/test/`      | `/app/test/dir1`          |
| `dir1`  | `/app/test`       | `/app/test`               |

## Use Cases

### 🚀 Full-Stack Deploy
Build backend + frontend, deploy to server, restart services — all in one command.

### 📊 Server Monitoring
Check disk usage, memory, service status across multiple servers instantly.

### 🗄️ Database Ops
Run backups, execute migrations, restore data — without SSH-ing manually.

### 🔒 SSL Management
Automate certificate renewal and Nginx reload.

### 🔄 Multi-Environment
Maintain dev / staging / production configs separately, deploy with `-c` flag.

### 📦 Release Pipeline
Chain tasks: test → build → deploy → health check — a complete CI/CD in one YAML.

## Release

```bash
./scripts/release.sh v0.1.0
```

Cross-compiles for all platforms and publishes a GitHub release.

## Feedback

Questions or contributions: koyeo@qq.com

## Contributing

Pull requests are welcome. For major changes, please open an issue first.

## License

[MIT](https://choosealicense.com/licenses/mit/)
