# Ship App — nest 部署功能的 TypeScript 重实现

> Created: 2026-08-28

## 结论

- 一句话方案：新增 `apps/ship`（npm 包 `@kozilla/ship`，bin `ship`），用 TypeScript 逐模块复刻 `apps/nest` 的 CLI 部署功能；配置文件 `ship.yaml`；**不实现** webui / `--ui` / `CommandEventHandler` / `HandlerPrompter`。
- 完成的可观测信号：
  - `pnpm --filter @kozilla/ship build` 产出 `dist/`，`ship --help` 列出 `init / list / run / storage / version` 5 个命令，`storage` 下有 `add / list / remove / usage / clean`。
  - 把 `apps/nest/nest-test.yaml` 改名为 `ship.yaml` 后 `ship run test` 3 个 step 全部执行、退出码 0。
  - 同一份 `deploy` 配置，nest 与 ship 对远端产生的目录结构一致，仅元数据目录名不同（`.nest/` → `.ship/`）。
  - domain 层（snapshot / conflict / backup / deploy service）的单测与 nest 的 4 个 `_test.go` 用例一一对应且通过。

## 约束（推导依据）

- nest 全部 Go 源 5.8k 行，其中 webui 500 行、`cmd/upload`（未注册到 root）350 行，剩余 ~4.9k 行是要复刻的范围。
- 功能面（来自 `apps/nest/cmd/nest.go` 与 `protocol/protocol.go`）：
  - 配置：`version / servers{host,port,user,password,identity_file,comment} / storages{alias→全局名} / envs / tasks{comment,workspace,envs,commands[]}`；command 四选一 `run | use | upload | deploy`；deploy = `servers[{use|inline}] + files[{source,target,storage?}] + commands[] + cwd + shell_init`。
  - 本地执行：`bash -c`，env = process.env ⊕ 全局 envs ⊕ task envs，cwd = task.workspace。
  - 远端执行：`bash -l -c '<单引号转义>'`，stdout/stderr 实时透传，stderr 保留 4096 字节尾部拼进错误。
  - 文件部署：本地 tar.gz → SFTP 写到 `targetDir/bundle-<name>.tar.gz~` 或经云存储 presigned URL 让远端 `curl` 下载到 `/tmp/nest-<name>.tar.gz` → 远端解压到 `targetDir/.nest/tmp` → 冲突分类（snapshot 管理的直接删；非管理的问用户 backup/remove）→ `mv` → 写 `targetDir/.nest/snapshot.json`（entries 按 bundle_name upsert，file 记 path/sha256/mtime）。
  - `target` 以 `/` 结尾视为目录，否则取父目录；`~` 开头或深度 <2 的绝对路径拒绝。
  - 云存储：对象 key = `nest/<sha1(bundle)>.tar.gz`；Head 比对 size 相同则跳过上传；同一次 run 内按本地路径去重；所有 server 下载完后批量删除本次上传的对象；OSS ≥32MiB 走分片(4MiB 起、≤10000 片、5 并发)，S3 单次 PutObject。
  - 用户级配置 `~/.nest/config.json`：`lang`、`encrypt_key`(base64 32B)、`storages{name→{provider,endpoint,region,bucket,access_key_id(密),access_key_secret(密)}}`；AES-256-GCM，密文 = base64(nonce‖ciphertext)。
  - `nest init`：写模板 `nest.yaml`，并把 `.nest` 注入 `.gitignore`。
- 运行环境：Node 24.19 / pnpm 10.28（已在 workspace 根安装）。
- 用户全局规则 G1/G2/G3：不写兜底、不写可选参数、不出现 `any/unknown/as`。nest 里存在多处兜底（见「未决」），移植时不能默默照搬。
- `yaml.parse()` 返回类型为 `any`，`JSON.parse()` 同理 —— 必须经运行时 schema 校验后才能得到类型化对象，否则违反 G3。

## 关键决策

- 语言 / 运行时：TypeScript ESM，Node ≥ 20 —— 包名 `@kozilla/ship` 决定了它走 npm 分发，不再有 Go 二进制。
- 依赖选型（按功能对应，不引入额外抽象层）：
  - CLI：`commander`（对应 cobra）。
  - YAML：`yaml`；schema 校验：`zod`（`z.infer` 直接给出类型，规避 `any/unknown`）。
  - SSH/SFTP：`ssh2`（同时提供 exec + sftp，对应 `golang.org/x/crypto/ssh` + `pkg/sftp`）。
  - tar.gz：`tar`（node-tar）。
  - OSS：`ali-oss`；S3：`@aws-sdk/client-s3` + `@aws-sdk/s3-request-presigner`。
  - 加密：`node:crypto` AES-256-GCM，密文格式与 nest 完全一致（nonce 12B ‖ ciphertext ‖ tag 16B，base64）。
  - 颜色：`picocolors`。
- 命名全部由 nest → ship：`ship.yaml`、`~/.ship/config.json`、本地临时目录 `.ship/tmp`、远端元数据目录 `<targetDir>/.ship/`（snapshot.json）、远端临时包 `/tmp/ship-<name>.tar.gz`、云对象前缀 `ship/`、日志前缀 `[ship]`。**不与 nest 共享任何状态**：同一台服务器上 nest 写的 `.nest/snapshot.json` 对 ship 不可见，首次用 ship 部署时那些文件会被当作"非管理文件"触发冲突询问。
- 不做 UI：删除 `webui/`、`--ui`、`CommandEventHandler`、`HandlerPrompter`、`SetContext/ctx` 取消链路（取消链路只服务于 UI 的停止按钮；CLI 下 Ctrl-C 由进程信号处理）。
- 保留 nest 的 DDD 分层：`deploy/domain`（纯函数 + 接口）、`deploy/application`（DeployService）、`deploy/infrastructure`（SSH 实现 + stdin prompter）。理由：4 个 `_test.go` 的用例依赖接口注入才能不连 SSH 就测通。
- 不移植 `cmd/upload`（nest 根命令已注释掉、未注册）和 `i18n` 的 zh/en 双份（nest 里两份文案完全相同，等价于单语）。`lang` 字段仍读写以保持 `config.json` 结构，但不再影响输出。
- 单测框架：`vitest`；只测 domain + application（与 nest 一致），infrastructure 不做 SSH 集成测试。
- 大文件 SFTP 上传：nest 每 1MiB `\r` 刷新进度行，照搬；bundle hash 用 stream 计算（nest 的 `server.go` 仍是 `ReadFile` 全读，`task.go` 已改 stream，ship 统一 stream）。

## 状态模型

- Snapshot（远端 `<targetDir>/.ship/snapshot.json`）：不存在 -> 存在（首次 deploy 成功）；entries[bundle_name] 每次 deploy 成功后被整体替换。
- 云对象（`ship/<sha1>.tar.gz`）：不存在 -> 已上传（本地 Head 与 size 不等）-> 已删除（同一 deploy step 内所有 server 下载完成后）。

## 未决 / 信息不足

以下是 nest 里的兜底/默认值，按 G1/G2 需要你拍板，否则 plan 不能开工。每条附推荐：

1. **server 默认值**：nest 在 port 为空时用 22，password/identity_file 都为空时用 `~/.ssh/id_rsa`。
   - A（推荐）：保留这两个默认，写入 zod schema 的 `.default()`，因为它们是 SSH 协议/OpenSSH 的既定约定，不是业务兜底。
   - B：port 必填、认证方式必填，缺失即 schema 报错。
2. **`~/.ship/config.json` 读取失败**：nest 任何读/解析错误都静默回退到默认配置。
   - A（推荐）：文件不存在 → 默认配置；文件存在但解析/schema 失败 → 报错退出（静默回退会让加密的 storages 凭空"消失"）。
   - B：完全照搬 nest，任何错误回退默认。
3. **冲突询问的默认输入**：nest 中直接回车 = backup，suffix 为空 = `.bak`。
   - A（推荐）：保留，这是交互提示里明示的默认值（`Choose [1]` / `[.bak]`），属于 UX 契约而非兜底。
   - B：必须显式输入，否则重问。
4. **远端 `.nest/` 兼容**：是否让 ship 同时读取远端旧的 `.nest/snapshot.json`（只读，用于识别 nest 时代部署的文件为"管理文件"）？
   - A（推荐）：不读。ship 是新工具，状态完全独立；旧文件首次部署走一次 backup 即可。
   - B：读旧 snapshot 作为 fallback 并迁移写入 `.ship/`。
5. **zod 边界**：`yaml.parse` / `JSON.parse` 返回 `any`，唯一消费点是 `schema.parse(...)`。是否接受"`any` 只出现在这一行的隐式传参、源码中不书写 `any` 字面量"作为 G3 的满足条件？
   - A（推荐）：接受。
   - B：不接受 → 需要自写 YAML/JSON 的 tokenizer，不现实。

## Plans 拆分

| 编号 | 标题 | 路径 | 依赖 | 状态 |
|---|---|---|---|---|
| 001 | Ship 包骨架、协议与存储命令 | `plans/001-ship-scaffold-protocol-storage.done.md` | - | 已完成 |
| 002 | Ship 执行器、部署域与 run 命令 | `plans/002-ship-executor-deploy-run.done.md` | 001 | 已完成 |
