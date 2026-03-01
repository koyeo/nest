# Configuration

Nest uses a single YAML file (`nest.yaml` by default) to define everything: servers, environment variables, and tasks.

## Config File Structure

```yaml
version: 1.0

servers:
  # ... server definitions

envs:
  # ... global environment variables

tasks:
  # ... task definitions
```

## Servers

Define your remote servers for deployment and remote command execution:

```yaml
servers:
  prod:
    comment: Production server        # Description
    host: 192.168.1.5                 # Server address
    port: 22                          # Port (default: 22)
    user: root                        # Username
    identity_file: ~/.ssh/id_rsa      # SSH key (default)
  staging:
    comment: Staging server
    host: 192.168.1.20
    user: deploy
    port: 2222
    password: mypassword              # Password auth (not recommended)
```

### Authentication

| Method | Field | Notes |
|:-------|:------|:------|
| SSH Key | `identity_file` | Default: `~/.ssh/id_rsa` |
| Password | `password` | Less secure, not recommended for production |

::: warning
Avoid committing password credentials to version control. Use SSH key auth whenever possible.
:::

## Environment Variables

Global environment variables are available to all task steps:

```yaml
envs:
  APP_NAME: myapp
  VERSION: "1.0.0"
  REMOTE_DIR: /opt/myapp
```

Task-level env vars can override globals:

```yaml
tasks:
  deploy:
    envs:
      BUILD_FLAGS: "-ldflags '-s -w'"
    steps:
      - run: go build $BUILD_FLAGS -o $APP_NAME .
```

## Tasks

Tasks are named groups of sequential steps:

```yaml
tasks:
  build:
    comment: Build the application     # Optional description
    workspace: ./frontend              # Optional working directory
    envs:                              # Task-scoped env vars
      NODE_ENV: production
    steps:
      - run: npm ci
      - run: npm run build
```

### Step Types

Each step in a task can be one of:

| Type | Description |
|:-----|:------------|
| `run` | Execute a local shell command |
| `deploy` | Upload files and/or execute remote commands |
| `upload` | Upload artifacts to cloud storage |
| `use` | Reference another task (task composition) |

### Task Composition with `use`

Reuse tasks to build pipelines:

```yaml
tasks:
  build:
    comment: Build backend
    steps:
      - run: go build -o myapp .

  deploy:
    comment: Deploy to server
    steps:
      - use: build              # Run "build" task first
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp

  release:
    comment: Full release pipeline
    steps:
      - run: go test ./...      # Test
      - use: deploy             # Build + Deploy
      - deploy:                 # Health check
          servers:
            - use: prod
          executes:
            - run: curl -sf http://localhost:8080/health
```

::: tip
Nest automatically detects circular `use` references and reports an error.
:::
