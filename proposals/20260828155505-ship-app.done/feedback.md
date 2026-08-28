# Feedback

执行 proposal 期间冒出的、未在当前会话处理的事项。收尾后由用户决定要不要新开 proposal / plan 处理。

---

## [plans/001-ship-scaffold-protocol-storage.md] YAML plain scalar 含 `: ` 的兼容性

- **类型**：设计调整
- **位置**：`apps/ship/src/protocol/load.ts`；`apps/nest/nest-test.yaml:6`
- **描述**：Go yaml.v3 接受 `run: echo "Step 1: Hello" && sleep 1`（plain scalar 内含 `: `），YAML 规范与 JS 生态所有解析器都拒绝。从 nest 迁移的 `nest.yaml` 若有此写法，ship 会报 `Nested mappings are not allowed in compact mappings`。
- **建议**：待决策 —— (a) 文档里说明需加引号或用 `|` 块；(b) 在报错信息里附加提示"把 run 值用单引号包裹或改用 `run: |`"。

## [plans/002-ship-executor-deploy-run.md] 云存储中继路径缺实机验证

- **类型**：范围外发现
- **位置**：`apps/ship/src/runner/cloud.ts`、`server-runner.ts#deployViaStorage`、`storage/{oss,s3}.ts`
- **描述**：本次无 OSS/S3 凭据，`files[].storage` 与 `upload:` 两条路径只有类型检查与代码对照，没有跑过真实上传 / presigned URL / 远端 curl / 收尾删除。ali-oss `head()` 的 content-length 读取、`signatureUrl` 参数、S3 `NotFound` 类型判断都值得一次真跑。
- **建议**：用 `ship storage add` 配一个测试 bucket 后跑一遍 `storage usage / clean` 与带 `storage:` 的 deploy。

## [plans/002-ship-executor-deploy-run.md] Ctrl-C 中断时远端命令不会被终止

- **类型**：设计调整
- **位置**：`apps/ship/src/execer/ssh-server.ts#pipeExec`
- **描述**：nest 的 ctx 取消链路（关闭 SSH session 中断远端命令）随 UI 一并移除。CLI 下 Ctrl-C 结束本进程，SSH 连接随之断开，远端 `bash -l` 通常收到 SIGHUP，但没有显式处理。
- **建议**：待决策 —— 如需确定性，在 `process.on("SIGINT")` 里主动 `ServerPool.close()`。
