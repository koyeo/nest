# CLI Reference

## Commands

### `nest init`

Initialize a configuration file and update `.gitignore`.

```bash
nest init              # Creates nest.yaml
nest init myconfig.yml # Creates custom config file
```

**What it does:**
- Creates the specified YAML config file with a starter template
- Adds `.nest`, `nest.yaml`, `nest.*.yml`, `nest.*.yaml` to `.gitignore`

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
```

**Step execution order:**
1. `run` — local shell command
2. `use` — invoke another task (supports circular dependency detection)
3. `upload` — compress and upload to cloud storage
4. `deploy` — upload files via SFTP and/or execute remote commands

---

### `nest list`

List all configured resources (servers, tasks, env vars).

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

---

### `nest version`

Print version information.

```bash
nest version
```

**Output:**
```
nest v0.2.0
  commit:  abc1234
  built:   2026-03-01T00:00:00Z
```

## Global Flags

| Flag | Short | Default | Description |
|:-----|:------|:--------|:------------|
| `--config <file>` | `-c` | `nest.yaml` | Specify config file path |
| `--help` | `-h` | — | Show help for any command |
