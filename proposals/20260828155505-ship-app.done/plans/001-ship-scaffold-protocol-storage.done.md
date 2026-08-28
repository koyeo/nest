# Ship 包骨架、协议与存储命令

> 来自 proposal: proposals/20260828155505-ship-app/

## 目标

- 交付可 build、可执行的 `@kozilla/ship` CLI，具备 `init / list / version / storage {add,list,remove,usage,clean}` 五组命令，以及供 002 使用的协议层（`ship.yaml` schema+load）、用户配置层（`~/.ship/config.json` + AES-GCM）、云存储层（OSS/S3 客户端）。

## 改动范围

- **新增** `apps/ship/`：
  - `package.json`（name `@kozilla/ship`，`bin.ship → dist/main.js`，`type: module`，`engines.node >=20`，scripts `build/dev/test/typecheck`）、`tsconfig.json`（`strict`、`noUncheckedIndexedAccess`、`exactOptionalPropertyTypes`）、`tsup.config.ts`（或 tsc）、`vitest.config.ts`。
  - `src/main.ts`：commander root，`-c/--config` 默认 `ship.yaml`，注册全部子命令；version 由 package.json 注入。
  - `src/common/const.ts`：`DEFAULT_CONFIG_FILE = "ship.yaml"`、`TMP_WORKSPACE = ".ship"`、`OBJECT_PREFIX = "ship/"`。
  - `src/protocol/schema.ts` + `load.ts`：zod schema 对应 nest `protocol.Config/Server/Task/Command/Upload/Deploy/FileMapping`；`Command` 用 4 选 1 的 union；`resolveStorage(config, alias)`。
  - `src/config/user-config.ts` + `crypto.ts`：`~/.ship/config.json` 读写、`generateEncryptKey / encrypt / decrypt`（与 nest 密文格式二进制兼容）、`addStorage / removeStorage / decryptStorage`。
  - `src/storage/{storage.ts,oss.ts,s3.ts,factory.ts}`：`ObjectStorage` 接口 `upload(key, filePath) / head(key) → size|-1 / listObjects(prefix) / deleteObjects(keys) / presignedUrl(key, expiresSeconds)`；OSS 分片参数照 nest（32MiB 阈值 / 4MiB 起 / 10000 片 / 5 并发）；批量删除按 1000 切片。
  - `src/utils/byte-size.ts`：对应 `unit.ByteSize`（输出 `12.5K` 风格）。
  - `src/logger.ts`：`step / print / error`，前缀 `[ship]`。
  - `src/cmd/{init,list,version,storage}.ts`：行为与 nest 对应命令逐条一致；`init` 模板改为 ship 词汇并注入 `.ship` 到 `.gitignore`；`storage add` 支持交互与 flag 两种模式。
  - 单测：`crypto`（往返 + 与一段 nest 生成的密文互解）、`schema`（合法/非法 yaml）、`byte-size`。
- **更新** 根 `package.json`：增加 `ship:build / ship:test / ship:dev` 脚本；`pnpm-workspace.yaml` 已含 `apps/*`，无需改。

## 验收

- [ ] `pnpm --filter @kozilla/ship build && node apps/ship/dist/main.js --help` 列出 5 个命令。
- [ ] `ship init` 在空目录生成 `ship.yaml` 与含 `.ship` 的 `.gitignore`；重复执行输出 `already exists` 且不改文件。
- [ ] `ship list -c <nest-test.yaml 改名后的文件>` 打印 version / tasks。
- [ ] `ship storage add x --provider oss --endpoint e --bucket b --access-key-id id --access-key-secret s` 后 `~/.ship/config.json` 中 `access_key_id` 为 base64 密文，`ship storage list` 显示掩码 `xxxx****xxxx`。
- [ ] 用 nest 的 `~/.nest/config.json` 里的 `encrypt_key` + 密文，ship 的 `decrypt` 能还原明文（格式兼容）。
- [ ] `pnpm --filter @kozilla/ship test` 与 `typecheck` 通过；`rg -n "any|unknown| as " apps/ship/src` 仅命中注释或字符串。

## 不变量

- 源码中不出现 `any` / `unknown` / `as` 类型断言（proposal 未决 5 的 A 选项通过后，`schema.parse(yaml.parse(text))` 是唯一 `any` 隐式传参点）。
- 所有函数参数必填（G2）；命令 flag 的"可选"由 commander 层处理，进入业务函数前已解析为确定值。
- `access_key_id / access_key_secret` 落盘前必须加密；`storage list` 永不打印明文。

## 关键点

- zod schema 里 `Command` 必须建模为互斥 union（`run | use | upload | deploy`），nest 用四个可空字段 + if/else 判断，直接照搬会引入大量可选字段。
- 未决 1/2 的答案决定 `Server` schema 是否带 `.default()` 与 `loadUserConfig` 的错误分支，开工前必须已拍板。
- ali-oss 的 `multipartUpload` 与 aws-sdk v3 的类型定义都较宽，接口返回值要在 `oss.ts / s3.ts` 内收窄成 `ObjectStorage` 声明的精确类型，不能把 SDK 类型泄漏到上层。

---

## 实施日志

- **执行时间**：2026-08-28 16:15
- **整体状态**：已完成

### 做了什么
- `apps/ship/`：package.json（`@kozilla/ship`，bin `ship`，ESM，node ≥20）、tsconfig（strict + noUncheckedIndexedAccess + exactOptionalPropertyTypes）、tsup、vitest、.gitignore。
- `src/common/const.ts`、`src/logger.ts`、`src/utils/byte-size.ts`、`src/utils/prompt.ts`。
- `src/protocol/schema.ts`（zod；`Command` 为 4 选 1 `.strict()` union）+ `load.ts`。
- `src/config/crypto.ts`（AES-256-GCM，nonce‖ct‖tag，base64）+ `user-config.ts`（`~/.ship/config.json`；文件不存在→默认，存在但非法→报错）。
- `src/storage/{storage,oss,s3,factory}.ts`（OSS 分片 32MiB/4MiB/10000/5；删除按 1000 切片）。
- `src/cmd/{init,list,version,storage}.ts` + `src/main.ts`；根 package.json 增加 `ship:build/test/dev`。
- 单测：byte-size、schema、crypto（8 个用例）。

### 验收核对
- [x] `build` 后 `--help` 列出 init / list / storage / version（+ commander 自带 help）；storage 下 add/list/remove/usage/clean。
- [x] `ship init` 生成 `ship.yaml` + 含 `.ship` 的 `.gitignore`；二次执行输出 `already exists`。
- [x] `ship list` 打印 version/tasks/servers。
- [x] `storage add x --provider oss ...` 后 config.json 中 access_key_id 为 base64 密文；`storage list` 显示 `fAvI****rZ4B`。
- [x] 用 nest 的 `config.Encrypt`（Go）生成密文，ship `decrypt` 还原 `hello-from-go`。
- [x] typecheck / 8 tests 通过；`grep -E "\bany\b|\bunknown\b| as [A-Z]" src` 无命中。

### 偏差与遗留
- `version: 1.0` 经 `yaml` 解析为数字 1，`list` 显示 `version: 1`（nest 显示 `1.0`）。仅展示差异。
- `nest-test.yaml` 中 `run: echo "Step 1: ..."` 是 YAML 规范不允许的 plain scalar（含 `: `），Go yaml.v3 宽容接受，`yaml` 与 `js-yaml` 均拒绝。ship 不做兼容；plan 002 验收用单引号包裹后的副本。已记 feedback.md。
- vitest 4 依赖 vite ≥6，而 workspace 内 docs 的 vitepress 提升了 vite 5，导致启动失败；在 apps/ship 显式加 `vite` devDependency 解决。
- `storage list/usage` 输出加了 `ship` 文案，`lang` 字段照 nest 读写但不再影响输出（i18n zh/en 文案本就相同）。
