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

1. **Upload** step: compress → compute SHA1 → upload to `nest/{project_hash}/{file}.tar.gz`
2. **Deploy** with `via`: generate pre-signed URL (1-hour expiry) → remote server downloads via `curl` → extract → deploy

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

### 2. Manage Storages

```bash
nest storage list              # List configured storages
nest storage remove oss-prod   # Remove a storage config
```

## Usage in `nest.yaml`

```yaml
tasks:
  deploy-overseas:
    comment: Build, upload to OSS, deploy via cloud storage
    steps:
      # Build locally
      - run: CGO_ENABLED=0 GOOS=linux go build -o myapp .

      # Upload artifact to cloud storage
      - upload:
          storage: oss-prod         # references global config
          source: ./myapp

      # Deploy via cloud storage (server downloads from OSS, not SFTP)
      - deploy:
          via: oss-prod             # download relay
          servers:
            - use: prod-us
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          executes:
            - run: systemctl restart myapp
```

## Supported Providers

| Provider | `--provider` | Notes |
|:---------|:-------------|:------|
| Alibaba Cloud OSS | `oss` | Requires endpoint (e.g. `oss-cn-hangzhou.aliyuncs.com`) |
| AWS S3 | `s3` | Requires region (e.g. `us-east-1`) |

::: tip
Storage credentials are stored in `~/.nest/config.json` with AES-256 encryption. They are never written to your project's `nest.yaml`.
:::
