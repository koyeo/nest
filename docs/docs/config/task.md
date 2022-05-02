---
sidebar_position: 3
---

# 任务配置

```yaml
version: 1.0
servers:
  server-1: # 定义服务器
    comment: 示例服务器
    host: 192.168.1.10
    user: root                                 # 默认使用 ~/.ssh/id_rsa 私钥进行认证
tasks:
  task-1: # 任务名称
    comment: 示例任务                           # 任务注释
    steps:
      - use: hi                                # 继承 hi 任务的 steps
      - run: go build -o foo foo.go            # 本地执行构建
      - deploy:
          servers:
            - use: server-1                    # 部署服务器  
          mappers:
            - source: ./foo                    # 本地文件路径
              target: /app/foo/bin/foo         # 服务器存放位置
          executes:
            - run: supervisorctl restart foo   # 服务器重启服务
      - run: rm foo                            # 本地清理
  hi:
    comment: 打个招呼
    steps:
      - run: echo "Hi! this is from nest~" 
```

## 部署文件映射

| source  | target            | 服务器存放位置               |
|:--------|:------------------|:----------------------|
| `file1` | `/app/test/file1` | `/app/test/file1`     |
| `file1` | `/app/test/file2` | `/app/test/file2`     |
| `file1` | **`/app/test`**   | `/app/test`           |
| `file1` | **`/app/test/`**  | `/app/test/file1`     |
| `dir1`  | `/app/test/dir2`  | `/app/test/dir2`      |
| `dir1`  | `/app/test/dir2/` | `/app/test/dir1/dir2` |
| `dir1`  | **`/app/test/`**  | `/app/test/dir1`      |
| `dir1`  | **`/app/test`**   | `/app/test`           |


