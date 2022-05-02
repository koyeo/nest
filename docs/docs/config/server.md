---
sidebar_position: 1
---

# 服务器

```yaml
version: 1.0
servers:
  server_1:                         # 服务器标识， 可以在 dploy 任务中，use 使用
    comment: 第一台服务器             # 备注
    host: 192.168.1.5               # 服务器地址
    port: 2222                      # 端口，默认使用 22
    user: root                      # 服务器用户名
    password: 123456                # 服务器密码，可以由 identity_file 选型替代
    identity_file: ~/.ssh/id_rsa    # 服务私钥认证文件， 默认使用 ~/.ssh/id_rsa
```