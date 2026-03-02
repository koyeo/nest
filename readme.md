<p align="center">
  <img src="./logo.png" alt="Nest" width="200" />
</p>

# Nest

A lightweight local CI/CD tool. One YAML, one command — build, deploy, and manage servers from your local machine.

## Quick Start

### Install

```bash
# Quick install (auto-detects OS & arch)
curl -fsSL https://raw.githubusercontent.com/koyeo/nest/master/scripts/install.sh | bash

# Or via Go
go install github.com/koyeo/nest@latest
```

### Usage

```bash
nest init                    # Create nest.yaml
nest run deploy              # Run a task
nest run deploy --raw        # Raw console output (no GUI)
nest run deploy -c prod.yml  # Use specific config
```

### Example `nest.yaml`

```yaml
version: 1.0

servers:
  prod:
    comment: Production server
    host: 192.168.1.10
    user: root

tasks:
  deploy:
    comment: Build and deploy
    steps:
      - run: go build -o myapp .
      - deploy:
          servers:
            - use: prod
          files:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          executes:
            - run: systemctl restart myapp
```

## Documentation

📖 Full docs: [docs/](./docs/)

- [Getting Started](./docs/guide/getting-started.md)
- [Configuration](./docs/guide/configuration.md)
- [Deployment](./docs/guide/deployment.md)
- [Cloud Storage Relay](./docs/guide/cloud-storage.md)
- [Multi-Environment](./docs/guide/multi-environment.md)
- [CLI Reference](./docs/reference/cli.md)

## License

[MIT](https://choosealicense.com/licenses/mit/)
