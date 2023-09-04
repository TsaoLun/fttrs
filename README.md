# 运行 Root 服务器

`go run . config_root.json`

# 运行 Sub1 服务器

`go run . config_sub1.json`

# 运行 Sub2 服务器

`go run . config_sub2.json`

# 测试

`curl 0.0.0.0:5500/sub1` 从 root 5500 端口访问 sub1 服务

`curl 0.0.0.0:5500/sub2` 从 root 5500 端口访问 sub2 服务

> 报错 `_cgo_export.c:3:10: fatal error: stdlib.h: No such file or directory #include <stdlib.h> ` 可尝试 `sudo apt install --reinstall build-essential`