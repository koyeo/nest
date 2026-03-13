# Deployment

Nest handles file uploads via SFTP and remote command execution over SSH — all defined in your `nest.yaml`.

## Deploy Step

A `deploy` step can upload files, run remote commands, or both:

```yaml
tasks:
  deploy:
    commands:
      - deploy:
          servers:
            - use: prod             # Reference a named server
          files:                     # File mappings (optional)
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          commands:                  # Remote commands (optional)
            - run: systemctl restart myapp
```

## Inline Servers

You can also define servers inline instead of referencing named ones:

```yaml
commands:
  - deploy:
      servers:
        - host: 192.168.1.10
          user: root
      files:
        - source: ./dist/
          target: /var/www/html/
```

## File Mapping Rules

The `source` → `target` mapping follows these rules:

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

::: tip
A trailing `/` on the target means "put the source **inside** this directory". No trailing `/` means "rename to this path".
:::

## Cloud Storage Relay

For large files or overseas servers, SFTP can be slow. Add `storage: <alias>` to a file mapping to transfer via cloud storage instead — Nest uploads to the bucket, then the remote server downloads via a pre-signed URL.

```yaml
storages:
  oss: oss-prod

tasks:
  deploy:
    commands:
      - deploy:
          servers:
            - use: prod
          files:
            # Cloud storage relay — upload to OSS, remote downloads via pre-signed URL
            - source: ./apps/web/.next/standalone/
              target: /root/app/
              storage: oss

            # Direct SFTP upload (default when storage is not set)
            - source: deploy/prod.toml
              target: /root/app/config.toml
```

::: tip Mixed Mode
You can mix storage-relayed and direct SFTP file mappings in the same deploy step.
:::

See [Cloud Storage Relay](./cloud-storage.md) for full setup instructions.

## Deploy Options

### `cwd` — Working Directory

Set a working directory for all `commands`:

```yaml
- deploy:
    servers:
      - use: prod
    cwd: /data/app
    commands:
      - run: npm install
      - run: pm2 restart app
```

### `shell_init` — Shell Initialization

Prepend an init command to every execute (e.g. loading nvm):

```yaml
- deploy:
    servers:
      - use: prod
    shell_init: source /root/.nvm/nvm.sh
    cwd: /data/app
    commands:
      - run: node -v
      - run: npm install
```

## Multi-Server Deploy

Deploy to multiple servers at once:

```yaml
tasks:
  deploy-all:
    commands:
      - deploy:
          servers:
            - use: prod
            - use: staging
          files:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          commands:
            - run: systemctl restart myapp
```

## Remote Commands Only

Skip file uploads and just run commands on remote servers:

```yaml
tasks:
  logs:
    comment: Tail production logs
    commands:
      - deploy:
          servers:
            - use: prod
          commands:
            - run: journalctl -u myapp -f --lines=100

  restart:
    comment: Restart service
    commands:
      - deploy:
          servers:
            - use: prod
          commands:
            - run: systemctl restart myapp
            - run: echo "✅ Service restarted"
```

## Common Use Cases

### Full-Stack Deploy

```yaml
tasks:
  deploy:
    comment: Build backend + frontend, deploy everything
    commands:
      - run: CGO_ENABLED=0 GOOS=linux go build -o myapp .
      - run: cd frontend && npm ci && npm run build
      - deploy:
          servers:
            - use: prod
          files:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
            - source: ./frontend/dist/
              target: /var/www/myapp/
          commands:
            - run: systemctl restart myapp
            - run: nginx -s reload
```

### Database Backup

```yaml
tasks:
  db-backup:
    comment: Backup production database
    commands:
      - deploy:
          servers:
            - use: prod
          commands:
            - run: |
                TIMESTAMP=$(date +%Y%m%d_%H%M%S)
                pg_dump -U postgres myapp_db > /opt/backups/myapp_${TIMESTAMP}.sql
                echo "✅ Backup saved: myapp_${TIMESTAMP}.sql"
```
