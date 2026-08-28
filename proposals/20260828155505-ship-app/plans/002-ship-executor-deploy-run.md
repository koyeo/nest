# Ship 执行器、部署域与 run 命令

> 来自 proposal: proposals/20260828155505-ship-app/

## 目标

- 交付 `ship run <task...>`：本地命令、`use` 组合、`upload` 到云存储、`deploy`（SFTP 直传 / 云存储中继 + 远端命令），行为与 nest 逐条一致，无 UI。

## 改动范围

- **新增** `apps/ship/src/`：
  - `execer/local-runner.ts`：`bash -c`，env 合并顺序 process.env → 全局 envs → task envs，cwd = task.workspace，stdin 透传，stderr 尾部 4096B 拼入错误。
  - `execer/ssh-server.ts` + `server-pool.ts`：基于 `ssh2` 的 `connect / exec(combined) / execPipe / sftp()`；命令统一包 `bash -l -c '<单引号转义>'`；password + identity_file 两种认证；`~` 展开；连接超时 60s；host key 不校验（与 nest 一致）；同 key 复用连接。
  - `utils/tar.ts`：目录/文件 → tar.gz，顶层前缀为源 basename；跳过损坏软链并在 stderr 打警告。
  - `deploy/domain/{snapshot,conflict,backup,interfaces}.ts`：`Snapshot.isManaged / addEntry`、`classifyConflicts`、`nextBackupName`、`RemoteFS / RemoteExec / UserPrompter / SnapshotRepository` 接口。
  - `deploy/application/deploy-service.ts`：解压到 `<targetDir>/.ship/tmp` → 列出顶层 → 冲突分类 → 管理文件删除 / 非管理文件询问 → `mv` → 计算远端 sha256 + mtime → 写 `<targetDir>/.ship/snapshot.json`。
  - `deploy/infrastructure/{ssh-remote-fs,ssh-remote-exec,snapshot-repo,stdin-prompter}.ts`：`ReadDir` 跳过 `.ship*`；`Remove` 对目录走 `rm -rf`；`FileHash` 用 `sha256sum || shasum -a 256`。
  - `runner/task-runner.ts`：`exec()` 顺序执行 commands；`use` 递归 + 环依赖检测（parents 集合）；`upload`/`uploadToStorage` 共用一套"压缩 → 流式 sha1(key)+sha256(hash) → Head 比 size → 上传"；`dedupKeys` 按本地路径去重；`deploy` 结束后 `cleanUploadedObjects` 按 alias 分组批量删。
  - `runner/server-runner.ts`：`upload(source, target)`（路径校验、targetDir 推导、SFTP 写 `bundle-<name>.tar.gz~`、每 1MiB 刷进度、`deployBundle`）、`pipeExec / combinedExec`；`deployFileViaStorage`：presigned URL 1h → 远端 `curl -fsSL` 到 `/tmp/ship-<name>.tar.gz` → `deployBundle` → `rm -f`。
  - `cmd/run.ts`：逐个 task 执行，缺 task 名或 task 不存在即报错退出 1；打印 start / success / failed。
  - 单测（vitest，内存 fake 实现 RemoteFS/RemoteExec/SnapshotRepository/UserPrompter）：移植 nest `backup_test / conflict_test / snapshot_test / snapshot_repo_test / deploy_service_test` 全部用例。

## 验收

- [ ] 把 `apps/nest/nest-test.yaml` 复制为 `ship.yaml`，`ship run test` 输出 3 个 step 且退出码 0；task 名不存在退出码 1。
- [ ] `use` 自引用（A use B, B use A）报 `circlely` 类错误而不是栈溢出。
- [ ] 对一台测试服务器 deploy 目录两次：第一次远端生成 `<target>/.ship/snapshot.json`；第二次同名文件被静默覆盖、不询问；手工放入一个未管理文件后第三次 deploy 出现 `[1] Backup / [2] Remove` 询问，选 1 生成 `<file>.bak`，再次冲突生成 `.bak.2`。
- [ ] `files[].storage: <alias>` 路径：云上出现 `ship/<sha1>.tar.gz`，deploy 完成后该对象被删除；两台 server 共用同一对象（日志出现 `already uploaded`）。
- [ ] `shell_init` + `cwd` 同时设置时远端命令形如 `cd <cwd> && <shell_init> && <cmd>`（与 nest 拼接顺序一致：先 shell_init 后 cd 包裹，即最终 `cd X && init && cmd`）。
- [ ] 单测全绿；`rg -n "any|unknown| as " apps/ship/src` 无类型级命中。

## 不变量

- 远端只在 `<targetDir>/.ship/` 与 `/tmp/ship-*.tar.gz` 写元数据/临时文件；临时文件在成功与失败路径都被清理。
- snapshot 中的文件在后续部署中永远不触发用户询问；不在 snapshot 中的同名文件永远触发询问（无 `--yes` 之类跳过项，nest 也没有）。
- 云对象删除只发生在同一 deploy step 内全部 server 下载完成之后。
- `target` 以 `~` 开头或绝对路径深度 <2（如 `/data`）必须被拒绝。

## 关键点

- `ssh2` 的 exec 流与 SFTP 句柄都是回调/事件式，要在 `ssh-server.ts` 内封成 Promise，并保证 stdout/stderr 全部 flush 后再 resolve/reject（nest 用 WaitGroup 保证这一点）。
- `bash -l -c` 单引号转义 `' → '"'"'` 必须逐字复刻，多行 `run: |` 依赖它。
- 进度行 `\rTotal: X Uploaded: Y` 用 `process.stdout.write`，不能走 logger（会换行）。
- 未决 3（冲突询问默认值）与未决 4（是否读旧 `.nest/`）决定 `stdin-prompter.ts` 与 `snapshot-repo.ts` 的分支，开工前必须已拍板。
