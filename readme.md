<p align="center">
  <img src="./logo.png" alt="Nest" width="200" />
</p>

# Nest

[中文文档](./README_zh.md)

A lightweight local CI/CD tool for rapid integration and deployment.

## Why Nest?

There are many CI/CD tools out there — Jenkins, GitHub Actions, Makefile, Ansible, Travis CI, etc. But they're often too heavy for quick delivery:

- **Jenkins** requires a full server deployment and web UI.
- **GitHub Actions / Travis CI** are platform-specific.
- **Ansible + Makefile** need to be configured together every time.

Nest solves these pain points with a single config file and CLI.

### Best For

- Solo full-stack developers who want fast feedback loops.
- Quick deployments of frontend/backend projects to servers.

### Not Ideal For

- Multi-person production environments requiring strict release management.

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

If you have Go installed:

```bash
go install github.com/koyeo/nest@latest
```

> **Note:** `go install` compiles and installs the `nest` binary to `$GOPATH/bin`. Make sure `$GOPATH/bin` is in your `$PATH`.

## Quick Start

### 1. Initialize Configuration

```bash
nest init
```

This will:
1. Create a `nest.yaml` file if it doesn't exist.
2. Add `.nest` to `.gitignore` to ignore the temporary workspace.

You can also specify a custom config file name:

```bash
nest init nest.production.yml
```

### 2. Edit `nest.yaml`

Here's an example that builds locally, deploys to a server, and restarts the service:

```yaml
version: 1.0
servers:
  server-1:
    comment: Example server
    host: 192.168.1.10
    user: root                                 # Uses ~/.ssh/id_rsa by default
tasks:
  task-1:                                      # Task name
    comment: Example task                      # Task description
    steps:
      - use: hi                                # Inherit steps from "hi" task
      - run: go build -o foo foo.go            # Build locally
      - deploy:
          servers:
            - use: server-1                    # Target server
          mappers:
            - source: ./foo                    # Local file
              target: /app/foo/bin/foo         # Remote destination
          executes:
            - run: supervisorctl restart foo   # Restart service on server
      - run: rm foo                            # Local cleanup
  hi:
    comment: Say hello
    steps:
      - run: echo "Hi! this is from nest~"
```

### 3. Run a Task

```bash
nest run task-1
```

Run multiple tasks:

```bash
nest run task-1 hi
```

## CLI Reference

### `nest init`

Initialize the `nest.yaml` config file and update `.gitignore`.

### `nest run <task...>`

Execute one or more tasks by name.

### `nest list`

List all configured resources including tasks, servers, and environment variables.

## Configuration Reference

### Servers

```yaml
servers:
  server_1:                         # Server identifier (used in deploy tasks)
    comment: My server              # Description
    host: 192.168.1.5               # Server address
    port: 2222                      # Port (default: 22)
    user: root                      # Username
    password: 123456                # Password (alternative to identity_file)
    identity_file: ~/.ssh/id_rsa    # Private key file (default: ~/.ssh/id_rsa)
```

### Environment Variables

```yaml
envs:
  k1: v1                            # Global key-value environment variables
  k2: v2
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

## Release

To create a new release:

```bash
./scripts/release.sh v0.1.0
```

This will run tests, cross-compile for all platforms (macOS/Linux/Windows × amd64/arm64), generate checksums, and publish a GitHub release.

## Feedback

For questions, contributions, or more information, reach out via email: koyeo@qq.com.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
