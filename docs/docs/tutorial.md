---
sidebar_position: 3
---

# 第一个任务


## 初始化配置

```bash
nest init
```

1. 如果目录下不存在 `nest.yml` 文件，则创建该文件。
2. 在 `.gitignore` 添加 `.nest` 行，以忽略 Nest 临时工作目录。


## 编辑 nest.yml
通过一些配置示例，实现如下功能：

1. 本地完成构建。
2. 将构建结果发布到服务器指定位置。
3. 在服务器执行重启。


```yml
version: 1.0
servers:
  server-1:
    comment: 示例服务器
    host: 192.168.1.10
    user: root                                 # 默认使用 ~/.ssh/id_rsa 私钥进行认证
tasks:
  task-1:                                      # 任务名称
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

## 执行任务

```
nest run task-1
```