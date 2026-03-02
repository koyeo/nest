# Deployment

Nest handles file uploads via SFTP and remote command execution over SSH — all defined in your `nest.yaml`.

## Deploy Step

A `deploy` step can upload files, run remote commands, or both:

```yaml
tasks:
  deploy:
    steps:
      - deploy:
          servers:
            - use: prod             # Reference a named server
          files:                     # File mappings (optional)
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          executes:                  # Remote commands (optional)
            - run: systemctl restart myapp
```

## Inline Servers

You can also define servers inline instead of referencing named ones:

```yaml
steps:
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

When source paths use a **storage protocol prefix** (e.g. `oss://`), Nest uploads the local file to cloud storage first, then the remote server downloads via a pre-signed URL. This is much faster than SFTP for overseas servers.

```yaml
storage:
  oss: oss-prod

tasks:
  deploy:
    steps:
      - deploy:
          servers:
            - use: prod
          files:
            # Cloud storage relay
            - source: oss://apps/web/.next/standalone/
              target: /root/app/
            # Direct SFTP upload
            - source: deploy/prod.toml
              target: /root/app/config.toml
```

See [Cloud Storage Relay](./cloud-storage.md) for full setup instructions.

## Multi-Server Deploy

Deploy to multiple servers at once:

```yaml
tasks:
  deploy-all:
    steps:
      - deploy:
          servers:
            - use: prod
            - use: staging
          files:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          executes:
            - run: systemctl restart myapp
```

## Remote Commands Only

Skip file uploads and just run commands on remote servers:

```yaml
tasks:
  logs:
    comment: Tail production logs
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: journalctl -u myapp -f --lines=100

  restart:
    comment: Restart service
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: systemctl restart myapp
            - run: echo "✅ Service restarted"
```

## Common Use Cases

### Full-Stack Deploy

```yaml
tasks:
  deploy:
    comment: Build backend + frontend, deploy everything
    steps:
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
          executes:
            - run: systemctl restart myapp
            - run: nginx -s reload
```

### Database Backup

```yaml
tasks:
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
```
