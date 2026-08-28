# @kozilla/ship

YAML-driven, local-first deploy CLI. One `ship.yaml` + one command: local build → SSH remote commands → file deploys over SFTP (with snapshot-based conflict handling) or via an OSS/S3 relay for slow links.

> **Using Claude Code?** This repo ships a [`use-kozilla-ship` skill](skills/use-kozilla-ship/SKILL.md) that teaches Claude Code how to author `ship.yaml`, wire cloud-storage relay, and run deploys. Copy it into your `.claude/skills/` and ask Claude to set up `ship`.

```bash
npm install -g @kozilla/ship
ship init          # scaffold ship.yaml
ship list          # show tasks / servers / envs
ship run deploy    # run a task
```

```yaml
version: 1.0
servers:
  prod: { host: 192.168.1.10, user: root, identity_file: ~/.ssh/id_rsa }
tasks:
  deploy:
    commands:
      - run: npm run build
      - deploy:
          servers: [ { use: prod } ]
          files:
            - source: ./dist
              target: /opt/app/
          commands:
            - run: systemctl restart app
```

Cloud relay: `ship storage encrypt oss --provider oss --endpoint … --bucket … --access-key-id … --access-key-secret …` prints a `storages:` block to paste into `ship.yaml`; set `storage: oss` on a `files:` entry. Credentials in `ship.yaml` are encrypted with a key built into ship — treat that bucket as public and clean it regularly (`ship storage clean oss`). If the storage is unavailable, ship logs it and falls back to SFTP.

Commands: `init`, `list`, `run <task...>`, `storage add|encrypt|list|remove|usage|clean`, `version`. Global flag `-c <file>` selects the config. Requires Node ≥ 20.
