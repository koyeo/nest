<p align="center">
  <img src="./logo.png" alt="Nest" width="200" />
</p>

# Nest

[English](./README.md)

适用于快速交付的本地集成和部署工具。对于中小型项目，Nest 可以完全替代传统运维工具 —— 构建、部署、服务器管理、运维操作，全部在本地一行命令搞定。

## 为什么选择 Nest？

CI/CD 的工具很多，比如 Jenkins、GitHub Actions、Makefile、Ansible、Travis CI 等，但它们太重了，不方便快速交付：

- **Jenkins** 要部署软件，通过网页使用。
- **GitHub Actions / Travis CI** 等平台性工具限定了使用场景，必须推送代码才能触发。
- **Ansible + Makefile** 二者要相互搭配，每次都要配一遍，很麻烦。

Nest 通过**一个 YAML 配置文件**和**一个命令行**解决这些痛点。

### 适用场景

- 独立或小团队全栈开发工程师，追求快速反馈。
- 快速将前后端项目发布到服务器上。
- 轻量级服务器运维：查日志、重启服务、数据库备份、SSL 证书续期。
- 多环境管理（dev / staging / production），通过不同配置文件隔离。

### 更优的选择

在多人协作的大型生产环境，可能需要严格的发版管理和审批流程，此时不建议使用 Nest。

## 安装

### 快速安装（推荐）

自动检测系统和架构，一行命令安装：

```bash
curl -fsSL https://raw.githubusercontent.com/koyeo/nest/master/scripts/install.sh | bash
```

如需**更新**到最新版本，重新执行上述命令即可。

### 通过 Go 安装

```bash
go install github.com/koyeo/nest@latest
```

> **注：** 请确保 `$GOPATH/bin` 已添加到 `$PATH` 路径下。

## 快速上手

### 1. 初始化配置

```bash
nest init
```

执行后将创建 `nest.yaml` 并在 `.gitignore` 中添加 `.nest`。

### 2. 编辑 `nest.yaml`

以下是一个实战示例 —— 构建 Go 后端、部署到服务器、重启服务：

```yaml
version: 1.0

servers:
  prod:
    comment: 生产服务器
    host: 192.168.1.10
    user: root
    # identity_file: ~/.ssh/id_rsa    # 默认 SSH 密钥
  staging:
    comment: 预发布服务器
    host: 192.168.1.20
    user: deploy
    port: 2222

envs:
  APP_NAME: myapp
  REMOTE_DIR: /opt/myapp

tasks:
  # ── 构建 & 部署 ──────────────────────────────────────
  deploy:
    comment: 构建并部署到生产环境
    steps:
      - run: echo "🔨 正在构建..."
      - run: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp .
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
            - source: ./configs/prod.yaml
              target: /opt/myapp/configs/app.yaml
          executes:
            - run: chmod +x /opt/myapp/bin/myapp
            - run: systemctl restart myapp
      - run: rm -f myapp
      - run: echo "✅ 部署完成！"

  # ── 前端部署 ──────────────────────────────────────────
  deploy-web:
    comment: 构建前端并部署静态文件
    steps:
      - run: cd frontend && npm ci && npm run build
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./frontend/dist/
              target: /var/www/myapp/
          executes:
            - run: nginx -s reload

  # ── 服务器运维 ────────────────────────────────────────
  logs:
    comment: 查看生产环境日志
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: journalctl -u myapp -f --lines=100

  status:
    comment: 检查所有服务器状态
    steps:
      - deploy:
          servers:
            - use: prod
            - use: staging
          executes:
            - run: systemctl status myapp && df -h && free -m

  restart:
    comment: 重启服务
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: systemctl restart myapp
            - run: echo "✅ 服务已重启"

  # ── 数据库 ────────────────────────────────────────────
  db-backup:
    comment: 备份生产数据库
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: |
                TIMESTAMP=$(date +%Y%m%d_%H%M%S)
                pg_dump -U postgres myapp_db > /opt/backups/myapp_${TIMESTAMP}.sql
                echo "✅ 备份完成: myapp_${TIMESTAMP}.sql"

  db-migrate:
    comment: 执行数据库迁移
    steps:
      - deploy:
          servers:
            - use: prod
          mappers:
            - source: ./migrations/
              target: /opt/myapp/migrations/
          executes:
            - run: cd /opt/myapp && ./bin/myapp migrate up

  # ── SSL 证书 ──────────────────────────────────────────
  cert-renew:
    comment: 续期 SSL 证书
    steps:
      - deploy:
          servers:
            - use: prod
          executes:
            - run: certbot renew --quiet && nginx -s reload

  # ── 完整发布流水线 ────────────────────────────────────
  release:
    comment: 完整发布流程 — 测试、构建、部署、验证
    steps:
      - run: go test ./...
      - use: deploy
      - deploy:
          servers:
            - use: prod
          executes:
            - run: curl -sf http://localhost:8080/health || exit 1
            - run: echo "✅ 健康检查通过"
```

### 3. 执行任务

```bash
nest run deploy            # 构建 & 部署
nest run logs              # 查看服务器日志
nest run status            # 检查服务器状态
nest run db-backup         # 备份数据库
nest run release           # 完整发布流水线
nest run deploy deploy-web # 执行多个任务
```

## 多环境管理 `--config`

通过 `-c` / `--config` 参数，使用不同的配置文件管理多个环境：

```bash
nest init                          # 创建 nest.yaml（默认 / 开发环境）
nest init nest.staging.yml         # 创建预发布配置
nest init nest.production.yml      # 创建生产配置
```

```bash
nest run deploy                    # 使用 nest.yaml（默认 / 开发环境）
nest run deploy -c nest.staging.yml       # 部署到预发布环境
nest run deploy -c nest.production.yml    # 部署到生产环境
nest list -c nest.production.yml          # 查看生产环境配置
```

通过不同的配置文件隔离环境，同时共享相同的任务定义。

## 云存储（OSS / S3）

通过 VPN 向海外服务器部署时，直接上传往往很慢。Nest 支持**云存储中继** —— 先将构建产物上传到 OSS/S3，然后让远程服务器通过预签名链接从云存储下载。

### 1. 添加云存储配置

**交互引导模式**（用户使用）：

```bash
nest storage add
```

按步骤引导：选择云服务商 → 输入 endpoint/region → 输入 bucket 名称 → 输入凭证。所有凭证**使用 AES-256 加密**存储在 `~/.nest/config.json` 中。

**命令行模式**（脚本 / AI 使用）：

```bash
nest storage add oss-prod \
  --provider oss \
  --endpoint oss-cn-hangzhou.aliyuncs.com \
  --bucket my-deploy-bucket \
  --access-key-id LTAI5t... \
  --access-key-secret xxxxxxxx
```

### 2. 管理云存储

```bash
nest storage list              # 列出已配置的云存储
nest storage remove oss-prod   # 删除云存储配置
```

### 3. 在 `nest.yaml` 中使用

```yaml
tasks:
  deploy-overseas:
    comment: 构建、上传到 OSS、通过云存储部署
    steps:
      # 本地构建
      - run: CGO_ENABLED=0 GOOS=linux go build -o myapp .

      # 上传构建产物到云存储
      - upload:
          storage: oss-prod         # 引用全局配置中的云存储
          source: ./myapp

      # 通过云存储部署（服务器从 OSS 下载，而非 SFTP）
      - deploy:
          via: oss-prod            # 下载中继
          servers:
            - use: prod-us
          mappers:
            - source: ./myapp
              target: /opt/myapp/bin/myapp
          executes:
            - run: systemctl restart myapp
```

**工作原理：**
1. `upload` 步骤：压缩 → 计算 SHA1 → 上传到 `nest/{项目哈希}/{文件名}.tar.gz`
2. `deploy` 带 `via`：生成预签名 URL（1 小时有效） → 远程服务器 `curl` 下载 → 解压 → 部署

## CLI 参考

| 命令 | 说明 |
|:-----|:-----|
| `nest init [file]` | 初始化配置文件并更新 `.gitignore` |
| `nest run <task...>` | 执行一个或多个任务 |
| `nest list` | 列出配置文件里的资源项 |
| `nest storage add [name]` | 添加云存储配置（交互式或通过参数） |
| `nest storage list` | 列出已配置的云存储 |
| `nest storage remove <name>` | 删除云存储配置 |

### 全局参数

| 参数 | 简写 | 说明 |
|:-----|:-----|:-----|
| `--config <file>` | `-c` | 指定配置文件路径（默认：`nest.yaml`） |

## 配置参考

### 服务器

```yaml
servers:
  my_server:
    comment: 我的服务器                # 备注
    host: 192.168.1.5                 # 服务器地址
    port: 2222                        # 端口（默认 22）
    user: root                        # 用户名
    password: 123456                  # 密码认证
    identity_file: ~/.ssh/id_rsa      # 密钥认证（默认）
```

### 环境变量

```yaml
envs:
  APP_NAME: myapp
  VERSION: "1.0.0"
```

### 部署文件映射

| source  | target            | 服务器存放位置               |
|:--------|:------------------|:----------------------|
| `file1` | `/app/test/file1` | `/app/test/file1`     |
| `file1` | `/app/test/file2` | `/app/test/file2`     |
| `file1` | `/app/test`       | `/app/test`           |
| `file1` | `/app/test/`      | `/app/test/file1`     |
| `dir1`  | `/app/test/dir2`  | `/app/test/dir2`      |
| `dir1`  | `/app/test/dir2/` | `/app/test/dir1/dir2` |
| `dir1`  | `/app/test/`      | `/app/test/dir1`      |
| `dir1`  | `/app/test`       | `/app/test`           |

## 使用场景

### 🚀 全栈部署
构建后端 + 前端，部署到服务器，重启服务 —— 一条命令搞定。

### 📊 服务器监控
即时检查多台服务器的磁盘、内存、服务状态。

### 🗄️ 数据库运维
执行备份、跑迁移、恢复数据 —— 无需手动 SSH 登录。

### 🔒 SSL 管理
自动化证书续期和 Nginx 热加载。

### 🔄 多环境隔离
维护 dev / staging / production 独立配置，通过 `-c` 参数切换。

### ☁️ 云存储中继
将构建产物上传到 OSS/S3，让远程服务器通过预签名 URL 下载 —— 绕过 VPN 慢速连接。

### 📦 发布流水线
串联任务：测试 → 构建 → 部署 → 健康检查 —— 一个 YAML 实现完整 CI/CD。

## 发版

```bash
./scripts/release.sh v0.1.0
```

交叉编译所有平台并发布 GitHub Release。

## 反馈

如果你有使用上的问题，或想参与项目的开发：koyeo@qq.com

## 贡献

欢迎提交 Pull Request。如果是较大的改动，请先创建 Issue 讨论。

## License

[MIT](https://choosealicense.com/licenses/mit/)
