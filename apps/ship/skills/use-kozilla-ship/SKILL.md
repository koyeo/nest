---
name: use-kozilla-ship
description: Use `ship` (`@kozilla/ship`) — a YAML-driven, local-first deploy CLI (TypeScript port of koyeo/nest). One shareable `ship.yaml` + one command drives local builds, remote command execution over SSH, file/directory deploys via SFTP with snapshot-based conflict handling, and cloud-storage relay (OSS/S3) with credentials embedded as `enc:` values (`ship storage encrypt`) and automatic SFTP fallback when the storage is unavailable. No UI mode. Use when a project has a `ship.yaml` (or `ship*.yml`), when the user wants to author/edit ship tasks, run a deploy, manage multi-environment configs, set up or share OSS/S3 relay credentials, migrate from `nest.yaml`, or asks how to use the `ship` command. Args (optional): a goal like "add a deploy task" or "set up OSS relay".
---

# use-kozilla-ship

`ship` is a YAML-driven task runner + deployment CLI. A single `ship.yaml` declares **servers**, **envs**, **storages** (cloud-relay aliases), and **tasks**; `ship run <task>` executes them locally and over SSH. Target audience: solo / small-team full-stack devs who want fast local-to-server delivery without Jenkins/Actions/Ansible.

Package: `@kozilla/ship` (npm), binary `ship`, Node ≥ 20. Source: `apps/ship` in the `ship` pnpm monorepo (github.com/koyeo/ship), next to the original Go `apps/nest`.

## Install

```bash
npm install -g @kozilla/ship          # global
npm install --save-dev @kozilla/ship  # or pinned per project, run via `npx ship …`
```

## When to reach for ship vs. not
- **Good fit:** quick deploys of front/back-end to a few servers, light ops (tail logs, restart services, DB backup/migrate, cert renew), multi-env isolation via separate config files, cloud-storage relay to bypass slow VPN links.
- **Not a fit:** large multi-person production with mandatory approval/release governance.

## CLI reference

| Command | Purpose |
|:--|:--|
| `ship init [file]` | Create a starter config (`ship.yaml` by default) and add `.ship` to `.gitignore` |
| `ship run <task...>` | Run one or more tasks **in order**; `ship run a b` runs a then b |
| `ship list` | Show tasks, servers, and envs from the config |
| `ship storage add [name]` | Add a cloud-storage config to `~/.ship/config.json` (interactive, or non-interactive with flags) |
| `ship storage encrypt [alias]` | Print a `storages:` block with credentials encrypted under ship's **built-in** key, to paste into a shareable `ship.yaml` |
| `ship storage list` | List configured storages (provider, bucket, endpoint) |
| `ship storage remove <name>` | Remove a storage config |
| `ship storage usage [name]` | Show object count + total size of ship artifacts in the bucket |
| `ship storage clean <name>` | Delete all ship-managed objects from the bucket (asks to confirm) |
| `ship version` | Print version |

**Global flag:** `-c, --config <file>` (default `ship.yaml`) — selects the config file; works on every subcommand.
**There is no `--ui` flag** — output is always raw terminal; interactive prompts (conflict resolution, `storage add`) read stdin.

Storage credentials live in one of two places (see Cloud storage relay): **inline in `ship.yaml`** as `enc:` values under ship's built-in key so the file is self-contained and shareable (default choice), or per-machine in `~/.ship/config.json` (AES-256-GCM, your own key) when the bucket must stay private.

## Config schema (source of truth: `apps/ship/src/protocol/schema.ts`)

```yaml
version: 1.0

servers:                          # named, referenced by `use:` in tasks
  prod:
    comment: Production server    # label shown in output
    host: 192.168.1.10
    port: 22                      # default 22
    user: root
    password: secret              # OR
    identity_file: ~/.ssh/id_rsa  # key auth (default when neither is set)

storages:                         # alias -> inline credentials (shareable) OR global config name (see Cloud Storage)
  oss:                            # inline: paste the block printed by `ship storage encrypt`
    provider: oss                 # oss | s3
    endpoint: oss-cn-hangzhou.aliyuncs.com   # OSS: endpoint; S3: region (+ optional endpoint)
    bucket: my-bucket
    access_key_id: enc:…          # must be `enc:` — plaintext keys are rejected at load time
    access_key_secret: enc:…
  oss-private: oss-prod           # global: name from `ship storage list` on this machine

envs:                             # global env vars (process.env ⊕ envs ⊕ task.envs)
  APP_NAME: myapp

tasks:
  deploy:
    comment: Build and deploy
    workspace: ./                 # cwd for local `run` steps (optional)
    envs: { FOO: bar }            # task-scoped envs (optional)
    commands:                     # executed top-to-bottom; each step is exactly ONE of run/use/upload/deploy
      - run: go build -o myapp .  # 1) run: local `bash -c` (supports YAML `|` multi-line)
      - use: other-task           # 2) use: invoke another task (cycles are detected)
      - upload:                   # 3) upload: compress+upload artifact to cloud storage
          storage: oss            #    alias from top-level `storages`
          source: ./myapp
      - deploy:                   # 4) deploy: push files and/or run remote commands
          servers:
            - use: prod           #    reference a named server (or inline host/user/...)
          cwd: /opt/myapp         #    remote commands run as `cd <cwd> && <shell_init> && <cmd>`
          shell_init: source ~/.nvm/nvm.sh
          files:
            - source: ./myapp           # local file or dir
              target: /opt/myapp/bin/myapp
              storage: oss              # optional: relay via cloud storage; omit = direct SFTP
          commands:
            - run: systemctl restart myapp   # remote, in `bash -l -c` (login shell → PATH/nvm available)
```

The schema is **strict**: a command step with two of `run/use/upload/deploy`, or an unknown key, is rejected at load time.

### YAML gotcha when migrating from `nest.yaml`
`ship` uses a spec-compliant YAML parser. A plain scalar containing `: ` is **invalid** even inside double quotes:
```yaml
- run: echo "Step 1: hello"      # ❌ "Nested mappings are not allowed in compact mappings"
- run: 'echo "Step 1: hello"'    # ✅ single-quote the whole value
- run: |                         # ✅ or use a block scalar
    echo "Step 1: hello"
```
Go's yaml.v3 (nest) accepted the first form; fix these when converting.

### source → target resolution (trailing slash matters)
`target` ending in `/` = directory, source lands inside it; otherwise the parent of `target` is the directory. `~…` targets and absolute paths shallower than 2 segments (`/data`) are rejected.

| source | target | lands at |
|:--|:--|:--|
| `file1` | `/app/test/file1` | `/app/test/file1` |
| `file1` | `/app/test/` | `/app/test/file1` |
| `dir1` | `/app/test/` | `/app/test/dir1` |
| `dir1` | `/app/test/dir2` | `/app/test/dir2` |

### Conflict handling on the remote (snapshot)
Each deploy writes `<targetDir>/.ship/snapshot.json` recording the files it placed. On the next deploy:
- files already in the snapshot (**managed**) are replaced silently;
- files **not** in the snapshot that would be overwritten trigger an interactive prompt: `[1] Backup (default, suffix `.bak`, then `.bak.2`, …)` / `[2] Remove`.
If stdin is closed (CI), the prompt fails with `read input error: stdin closed` and the deploy aborts — deploy into an empty/managed directory in non-interactive runs.
State is **not shared with nest**: a server previously deployed by nest (`.nest/snapshot.json`) is treated as unmanaged on ship's first deploy — expect one backup prompt.

## Cloud storage relay (OSS / S3)
Bypasses slow direct SFTP: local uploads the tarball to OSS/S3 (object key `ship/<sha1>.tar.gz`, skipped if the same size already exists), the remote `curl`s it via a 1-hour pre-signed URL, extracts, and the objects are deleted after all servers finished.

**Automatic fallback:** if the storage cannot be resolved/constructed, or upload → presign → remote `curl` fails, ship logs `⚠️ storage '<alias>' unavailable: …` + `↩️ … falling back to direct SFTP upload` and transfers that file over SFTP instead. Failures *after* the bundle reached the remote (extract/conflict/snapshot) are not retried. A pure `upload:` step has nothing to fall back to and errors out.

Two ways to provide credentials — default to **A** so `ship.yaml` stays self-contained; use **B** only when the bucket must stay private.

**A. Inline in `ship.yaml` (`ship storage encrypt`) — default, shareable.** The key is baked into the ship binary, so this is obfuscation, not secrecy.
1. `ship storage encrypt oss --provider oss --endpoint oss-cn-hangzhou.aliyuncs.com --bucket my-bucket --access-key-id … --access-key-secret …` (S3: `--region` instead of `--endpoint`; omit flags for the wizard). It prints a ready `storages:` block — paste it into `ship.yaml`.
2. Every run that touches an inline alias prints: *credentials are embedded in ship.yaml — anyone holding this file can read/write the bucket, treat it as PUBLIC* and reminds `ship storage clean <alias>`. Consequences to bake into the setup:
   - use a **dedicated throw-away bucket** for ship artifacts only; never point it at a bucket holding real data;
   - the bucket only ever holds `ship/<sha1>.tar.gz` build bundles, and ship deletes them after each deploy — but interrupted runs leave objects behind, so **schedule `ship storage clean <alias>`** (or check `ship storage usage <alias>`) regularly.

**B. Per-machine (`ship storage add`) — private.** Credentials stay in `~/.ship/config.json`; `ship.yaml` only carries the alias → name mapping, and every teammate must run `ship storage add` once.
1. `ship storage list`. If the storage is missing: `ship storage add oss-prod --provider oss --endpoint … --bucket … --access-key-id … --access-key-secret …`.
2. `storages: { oss: oss-prod }` — value must match a name from `ship storage list`.

Then set `storage: <alias>` on a `files:` mapping (or an `upload:` step). Mixed mode is fine; with multiple servers each source uploads once. `ship storage usage/clean <alias>` accept both kinds of alias (pass `-c` if the config file isn't `ship.yaml`).

**Before authoring a task that transfers files: read `storages:` in `ship.yaml`, then `ship storage list`.** If either has a usable storage, route the transfer through it (`storage: <alias>`); the SFTP fallback means a relay entry never makes a deploy fail harder than plain SFTP would.

## Multi-environment
```bash
ship init ship.production.yml
ship run deploy -c ship.production.yml
ship list -c ship.production.yml
```

## Common recipes
```bash
ship init                       # scaffold ship.yaml (+ .ship in .gitignore)
ship storage encrypt oss …      # → paste block into ship.yaml `storages:`
ship list                       # confirm the config parses
ship run deploy                 # build & deploy
ship run test deploy health     # chained pipeline, in order
ship run deploy -c ship.staging.yml
ship storage usage oss          # what's left in the bucket
ship storage clean oss          # wipe ship/ objects (asks for "yes")
```

## Setting up a project from scratch (order matters)
1. `ship init` → edit `servers:` (prefer `identity_file` over `password`).
2. If deploys go over a slow/VPN link: `ship storage encrypt <alias> …` → paste into `storages:`; otherwise skip and let `files:` use direct SFTP.
3. Write tasks: a local `build` task, and a `deploy` task that `use: build`s then has one `deploy:` step per environment/server group.
4. `ship list`, then `ship run build` alone, then the full `ship run deploy` — the first deploy into a non-empty target directory will prompt for backup/remove of unmanaged files, so run it interactively.

## Working in this kind of project
- **Before writing any task that moves files: run `ship storage list`.** Prefer `storage: <alias>` on the `files:` mapping over direct SFTP on slow/VPN links.
- Edit `ship.yaml`, then `ship list` to confirm it parses (strict schema) and to see resolved tasks/servers.
- Prefer named `servers` + `- use:` over inline duplication.
- Storage credentials: inline `enc:` (via `ship storage encrypt`) by default so the file is shareable — on a throw-away bucket you treat as public; per-machine `ship storage add` only when the bucket must stay private.
- Validate a config change with a `run: echo` step before wiring real remote commands.
- Converting from nest: rename `nest.yaml` → `ship.yaml`, fix `: `-in-plain-scalar values (see gotcha), turn each `storages:` entry into an inline block with `ship storage encrypt` (nest's `~/.nest/config.json` is not read), drop `--ui`, and expect one backup prompt on the first deploy to a nest-managed directory.

## Working on ship itself (the `koyeo/ship` monorepo)
- Layout: `apps/ship` (TypeScript, tsup → `dist/main.js`, vitest), `apps/nest` (Go original), `docs` (VitePress). Root scripts: `pnpm ship:build | ship:test | ship:dev | ship:publish`.
- Rules baked into the codebase: strict TS (`noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`), no `any` / `unknown` / `as` casts, no optional function parameters; YAML/JSON enter the program only through zod schemas (`src/protocol/schema.ts`, `src/config/user-config.ts`, `src/deploy/domain/snapshot.ts`).
- Feature map: `src/cmd/*` (commander commands) → `src/runner/task-runner.ts` (step loop, `use` cycle check, SFTP fallback) → `src/runner/server-runner.ts` (SFTP upload / cloud relay, `StorageUnavailableError`) → `src/deploy/application/deploy-service.ts` (extract → conflict → move → snapshot) with `src/deploy/infrastructure/*` (ssh2 adapters, stdin prompter). Cloud: `src/runner/cloud.ts` + `src/storage/{oss,s3}.ts`. Built-in inline key: `src/config/inline-crypto.ts`.
- Tests: `pnpm -F @kozilla/ship test` (domain / deploy-service / snapshot-repo / schema / crypto, in-memory fakes in `src/deploy/test-fakes.ts`); SSH paths are verified manually against a throw-away `alpine` sshd container (`docker run -p 2222:22 …`), not in CI.
- Release: `./scripts/publish.sh [X.Y.Z]` at the repo root — typecheck + build + test, bump `apps/ship/package.json`, rebuild, `pnpm publish --access public`, commit `chore(ship): release vX.Y.Z`, tag `ship-vX.Y.Z` (the bare `v*` tags belong to nest's Go releases). It does **not** push; run `git push && git push --tags` afterwards.

