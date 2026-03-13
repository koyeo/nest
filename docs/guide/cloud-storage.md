# Cloud Storage Relay

When deploying to overseas servers via VPN, direct SFTP uploads can be very slow. Nest supports **cloud storage relay** — upload build artifacts to an OSS/S3 bucket, then have the remote server download from the cloud using a pre-signed URL.

## How It Works

```
Local Machine                    Cloud Storage (OSS/S3)              Remote Server
┌──────────┐    upload           ┌──────────────────┐    curl        ┌──────────┐
│  Build   │ ────────────────>   │  nest/{hash}/    │ <──────────── │  Deploy  │
│ Artifact │    compress+sha1    │  myapp.tar.gz    │  pre-signed   │  Server  │
└──────────┘                     └──────────────────┘    URL (1h)    └──────────┘
```

1. Nest detects the `storage` field on a file mapping entry
2. Compresses the local path → computes SHA1 hash → uploads to `nest/{hash}.tar.gz`
3. Generates a pre-signed URL (1-hour expiry)
4. Remote server downloads via `curl` → extracts to target directory

## Setup

### 1. Add a Storage Config

**Interactive mode** (guided):

```bash
nest storage add
```

Follows a step-by-step guide: provider → endpoint/region → bucket name → credentials. All credentials are **AES-256 encrypted** and stored in `~/.nest/config.json`.

**Non-interactive mode** (for scripts / automation):

```bash
nest storage add oss-prod \
  --provider oss \
  --endpoint oss-cn-hangzhou.aliyuncs.com \
  --bucket my-deploy-bucket \
  --access-key-id LTAI5t... \
  --access-key-secret xxxxxxxx
```

### 2. Declare in `nest.yaml`

Map a storage alias to the global config name:

```yaml
storages:
  oss: oss-prod          # "oss" is the alias, "oss-prod" is the global config name
```

### 3. Manage Storages

```bash
nest storage list              # List configured storages
nest storage remove oss-prod   # Remove a storage config
```

## Usage in `nest.yaml`

Set the `storage` field on any file mapping to route it through cloud storage:

```yaml
storages:
  oss: oss-prod

tasks:
  deploy:
    comment: Build and deploy via cloud storage
    commands:
      - run: pnpm run build

      - deploy:
          servers:
            - use: prod
          files:
            # storage: oss → upload to OSS, remote downloads via pre-signed URL
            - source: ./apps/web/.next/standalone/
              target: /root/app/
              storage: oss
            - source: ./apps/web/.next/static
              target: /root/app/.next/static
              storage: oss

            # No storage field → direct SFTP upload
            - source: deploy/prod.toml
              target: /root/app/config.toml
          commands:
            - run: pm2 restart web
```

::: tip Mixed Mode
You can mix storage-relayed and direct SFTP file mappings in the same deploy step.
File mappings without a `storage` field use direct SFTP transfer.
:::

::: tip Automatic Dedup
When deploying to multiple servers, Nest uploads each source to cloud storage only once. Subsequent servers reuse the same pre-signed URL.
:::

## Supported Providers

| Provider | `--provider` | Notes |
|:---------|:-------------|:------|
| Alibaba Cloud OSS | `oss` | Requires endpoint (e.g. `oss-cn-hangzhou.aliyuncs.com`) |
| AWS S3 | `s3` | Requires region (e.g. `us-east-1`) |

::: tip
Storage credentials are stored in `~/.nest/config.json` with AES-256 encryption. They are never written to your project's `nest.yaml`.
:::
