# CLI Reference

## Commands

### `nest init`

Create a starter config file and update `.gitignore`.

```bash
nest init              # Creates nest.yaml
nest init myconfig.yml # Creates custom config file
```

**What it does:**
- Creates the specified YAML config file with a starter template
- Adds `.nest` to `.gitignore`

---

### `nest run`

Execute one or more tasks defined in the config file.

```bash
nest run <task...>
```

**Examples:**
```bash
nest run deploy              # Run single task
nest run build deploy        # Run multiple tasks in order
nest run deploy -c prod.yml  # Use specific config file
nest run deploy --ui         # Use web UI with step tree and live output
```

**Flags:**

| Flag | Default | Description |
|:-----|:--------|:------------|
| `--ui` | `false` | Launch a web-based UI with command tree and live output |

By default, `nest run` outputs directly to the terminal (raw mode). Use `--ui` to open a visual web interface.

**Command execution order:**
1. `run` — local shell command (supports multi-line YAML `|` blocks)
2. `use` — invoke another task (supports circular dependency detection)
3. `upload` — compress and upload artifacts to cloud storage
4. `deploy` — upload files and/or execute remote commands

**Deploy step options:**

| Field | Description |
|:------|:------------|
| `servers` | Target servers (by name reference or inline) |
| `files` | File mappings with `source`, `target`, and optional `storage` |
| `commands` | Commands to run on each server after upload |
| `cwd` | Working directory for all commands |
| `shell_init` | Init command prepended to each command (e.g. `source ~/.nvm/nvm.sh`) |

**File mapping fields:**

| Field | Required | Description |
|:------|:---------|:------------|
| `source` | Yes | Local file or directory path |
| `target` | Yes | Remote destination path |
| `storage` | No | Storage alias name for cloud relay; empty = direct SFTP |

---

### `nest list`

Display all configured tasks, servers, and environment variables.

```bash
nest list
nest list -c nest.production.yml
```

---

### `nest storage`

Manage cloud storage configurations for OSS/S3 relay.

#### `nest storage add`

```bash
# Interactive mode
nest storage add

# Non-interactive mode
nest storage add <name> \
  --provider <oss|s3> \
  --endpoint <endpoint> \
  --bucket <bucket-name> \
  --access-key-id <key-id> \
  --access-key-secret <secret>
```

#### `nest storage list`

```bash
nest storage list
```

#### `nest storage remove`

```bash
nest storage remove <name>
```

#### `nest storage usage`

```bash
nest storage usage <name>    # Show object count and total size
```

#### `nest storage clean`

```bash
nest storage clean <name>    # Delete all nest objects (with confirmation)
```

---

### `nest version`

Print version, commit hash, and build time.

```bash
nest version
```

## Global Flags

| Flag | Short | Default | Description |
|:-----|:------|:--------|:------------|
| `--config <file>` | `-c` | `nest.yaml` | Specify config file path |
| `--help` | `-h` | — | Show help for any command |

