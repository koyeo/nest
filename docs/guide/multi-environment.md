# Multi-Environment

Use the `-c` / `--config` flag to manage different environments with separate config files.

## Setup

Create config files for each environment:

```bash
nest init                          # Creates nest.yaml (default / dev)
nest init nest.staging.yml         # Create staging config
nest init nest.production.yml      # Create production config
```

## Usage

```bash
# Development (default)
nest run deploy

# Staging
nest run deploy -c nest.staging.yml

# Production
nest run deploy -c nest.production.yml

# View production config
nest list -c nest.production.yml
```

## Example Structure

```
project/
├── nest.yaml                # Dev config (default)
├── nest.staging.yml         # Staging config
├── nest.production.yml      # Production config
└── ...
```

### `nest.yaml` (Dev)

```yaml
version: 1.0

servers:
  dev:
    host: 192.168.1.100
    user: dev

tasks:
  deploy:
    commands:
      - run: go build -o myapp .
      - deploy:
          servers:
            - use: dev
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
```

### `nest.production.yml` (Production)

```yaml
version: 1.0

servers:
  prod:
    host: 10.0.0.50
    user: deployer
    port: 2222

tasks:
  deploy:
    commands:
      - run: CGO_ENABLED=0 GOOS=linux go build -ldflags '-s -w' -o myapp .
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          commands:
            - run: systemctl restart myapp
```

::: tip
All `nest.yaml`, `nest.*.yml`, and `nest.*.yaml` files are automatically added to `.gitignore` by `nest init` — keeping your server credentials out of version control.
:::
