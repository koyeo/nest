# Nest

适用于快速交付的本地集成部署工具。

## 安装

```bash
go install github.com/koyeo/nest
```

## 项目初始化

```bash
nest init
```

1. 如果目录下不存在 `nest.yml` 文件，则创建该文件。
2. 在 `.gitignore` 添加 `.nest` 行，以忽略 Nest 临时工作目录。

## 第一个工作流

通过一些配置示例，实现如下功能：

1. 本地完成构建。
2. 将构建结果发布到服务器指定位置。
3. 在服务器执行重启。

**编辑 nest.yml：**

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

**执行工作流：**

```
nest run task-1
```

更多用法参见文档：[https://nest.kozilla.io](https://nest.kozilla.io)。

## 贡献
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)
