# Getting Started

## Why Nest?

There are many CI/CD tools — Jenkins, GitHub Actions, Makefile, Ansible, Travis CI — but they're often too heavy for quick delivery:

- **Jenkins** requires a full server deployment and web UI.
- **GitHub Actions / Travis CI** are platform-specific and require pushing code to trigger.
- **Ansible + Makefile** need to be configured together every time.

Nest solves these pain points with a **single YAML config** and a **single CLI command**.

### Best For

- Solo or small-team full-stack developers who want fast feedback loops.
- Quick deployments of frontend/backend projects to servers.
- Lightweight server ops: log checks, service restarts, DB backups, cert renewals.
- Multi-environment management (dev / staging / production) with separate config files.

## Installation

### Quick Install (Recommended)

Auto-detects your OS and architecture:

```bash
curl -fsSL https://raw.githubusercontent.com/koyeo/nest/master/scripts/install.sh | bash
```

To **update**, run the same command again.

### Install via Go

```bash
go install github.com/koyeo/nest@latest
```

::: tip
Make sure `$GOPATH/bin` is in your `$PATH`.
:::

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

envs:
  APP_NAME: myapp

tasks:
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
          executes:
            - run: chmod +x /opt/myapp/bin/myapp
            - run: systemctl restart myapp
      - run: rm -f myapp
      - run: echo "✅ Deployed!"
```

### 3. Run Tasks

```bash
nest run deploy            # Build & deploy
nest run logs              # Tail server logs
nest run status            # Check server status
nest run deploy deploy-web # Run multiple tasks
```

## Next Steps

- [Configuration Reference →](/guide/configuration)
- [Deployment Guide →](/guide/deployment)
- [Cloud Storage Relay →](/guide/cloud-storage)
- [Multi-Environment →](/guide/multi-environment)
- [CLI Reference →](/reference/cli)
