# 帧桥 API 二开指南

帧桥 API 是基于 One API fork 的视频生成 API 中转站原型。当前阶段先复用 One API 已有的用户、令牌、渠道、额度、日志、兑换码和管理后台能力；后续围绕 Seedance 视频生成补齐异步任务、预扣费、结算、失败退款、账本、对象存储和 webhook。

## 相关文档

- [项目实施进度跟踪](./framebridge-implementation-progress.md)
- [One API 当前架构文档](./framebridge-one-api-architecture.md)
- [One API 管理 API 文档](./API.md)

## 本地架构

本地二开推荐使用“依赖服务跑 Docker，业务服务跑源码”的方式：

```text
浏览器 / API Client
  -> Go 服务，端口 3000
  -> MySQL，Docker 端口 13306
  -> Redis，Docker 端口 16379
```

主要目录：

- `main.go`：Go 服务入口，会读取 `.env` 并启动 Gin HTTP 服务。
- `router/`、`controller/`、`relay/`：后端路由、管理接口和模型中转逻辑。
- `model/`：GORM 数据模型和自动迁移逻辑。
- `web/default/`：React 管理台源码。
- `web/build/default/`：前端构建产物，Go 服务通过 `go:embed` 嵌入这里的静态文件。

注意：修改前端后必须先重新 `npm run build`，再重新启动或重新编译 Go 服务，否则页面仍然会使用旧的嵌入资源。

## 本地启动

前置依赖：

- Go
- Node.js / npm
- Docker CLI + Docker Compose
- macOS 上可使用 Colima 提供 Docker Engine

启动 Docker Engine：

```bash
colima start
```

启动 MySQL 和 Redis：

```bash
cd /Users/bytedance/my_workspace/VideoBridge
docker compose up -d db redis
docker compose ps
```

本地开发配置放在 `.env`，可以从 `.env.example` 复制：

```bash
cp .env.example .env
```

当前本地约定端口：

```env
PORT=3000
SQL_DSN=oneapi:123456@tcp(127.0.0.1:13306)/one-api
REDIS_CONN_STRING=redis://127.0.0.1:16379
SYNC_FREQUENCY=60
SESSION_SECRET=framebridge-local-dev-secret
TZ=Asia/Shanghai
```

首次启动前构建前端：

```bash
cd /Users/bytedance/my_workspace/VideoBridge/web/default
npm install --legacy-peer-deps
npm run build
```

启动源码服务：

```bash
cd /Users/bytedance/my_workspace/VideoBridge
go run main.go --log-dir ./logs
```

打开：

```text
http://localhost:3000
```

默认管理员账号：

```text
root / 123456
```

首次登录后必须修改默认密码。

如果需要后台常驻运行，可以用 `screen`：

```bash
cd /Users/bytedance/my_workspace/VideoBridge
go build -o /tmp/framebridge/framebridge-dev-server main.go
screen -dmS framebridge-dev zsh -lc "cd /Users/bytedance/my_workspace/VideoBridge && /tmp/framebridge/framebridge-dev-server --log-dir ./logs > logs/framebridge-dev.log 2>&1"
```

查看和停止：

```bash
screen -ls
tail -f logs/framebridge-dev.log
screen -S framebridge-dev -X quit
```

## 开发流程

后端开发：

```bash
cd /Users/bytedance/my_workspace/VideoBridge
go run main.go --log-dir ./logs
```

只改 Go 代码时，重启 Go 服务即可。数据库表结构由 `model.InitDB()` 中的 GORM `AutoMigrate` 自动迁移。

前端开发：

```bash
cd /Users/bytedance/my_workspace/VideoBridge/web/default
npm run build

cd /Users/bytedance/my_workspace/VideoBridge
go run main.go --log-dir ./logs
```

前端依赖如果出现 `Cannot find module 'ajv/dist/compile/codegen'`，安装兼容版本：

```bash
cd /Users/bytedance/my_workspace/VideoBridge/web/default
npm install --save-dev ajv@8.17.1 --legacy-peer-deps
```

常用配置：

- `.env`：本地源码启动使用，已被 `.gitignore` 忽略。
- `docker-compose.yml`：本地依赖服务和 Docker demo 使用。
- `SESSION_SECRET`：本地可用固定字符串，生产必须换成强随机值。
- `SQL_DSN`：不设置会退回 SQLite；二开建议固定使用 MySQL。
- `REDIS_CONN_STRING` + `SYNC_FREQUENCY`：同时设置后 Redis 才会启用。

## 测试流程

依赖服务检查：

```bash
docker compose ps
mysql -h127.0.0.1 -P13306 -uoneapi -p123456 -e "select database(), current_user();" one-api
redis-cli -h 127.0.0.1 -p 16379 ping
```

后端健康检查：

```bash
curl -sS http://localhost:3000/api/status
```

前端静态资源检查：

```bash
curl -sS http://localhost:3000/ | head
ASSET_PATH=$(curl -sS http://localhost:3000/ | sed -n 's/.*src="\([^"]*main\.[^"]*\.js\)".*/\1/p')
curl -I "http://localhost:3000${ASSET_PATH}"
```

构建检查：

```bash
cd /Users/bytedance/my_workspace/VideoBridge
go test ./...

cd /Users/bytedance/my_workspace/VideoBridge/web/default
npm run build
```

业务冒烟测试：

1. 登录后台并修改 `root` 默认密码。
2. 在「渠道」里新增一个上游渠道。
3. 在「令牌」里创建测试 Token。
4. 调用 OpenAI 兼容接口：

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <你的 One API 测试 Token>" \
  -d '{
    "model": "<后台渠道支持的模型名>",
    "messages": [
      {
        "role": "user",
        "content": "hello"
      }
    ]
  }'
```

## 部署流程

Demo 部署可以直接使用 Docker Compose。当前 `docker-compose.yml` 默认使用官方 One API 镜像，适合快速验证后台和基础中转能力：

```bash
cd /Users/bytedance/my_workspace/VideoBridge
docker compose up -d
docker compose ps
```

二开后的源码部署流程：

```bash
cd /path/to/VideoBridge/web/default
npm install --legacy-peer-deps
npm run build

cd /path/to/VideoBridge
go mod download
go build -ldflags "-s -w" -o framebridge-api main.go

SQL_DSN='oneapi:<password>@tcp(<mysql-host>:3306)/one-api' \
REDIS_CONN_STRING='redis://<redis-host>:6379' \
SYNC_FREQUENCY='60' \
SESSION_SECRET='<strong-random-string>' \
TZ='Asia/Shanghai' \
./framebridge-api --port 3000 --log-dir ./logs
```

生产注意事项：

- 修改 `root / 123456` 默认密码。
- 修改 MySQL 密码和 `SESSION_SECRET`。
- 使用 MySQL，不要依赖 SQLite 承载生产数据。
- 只对外暴露 Web 服务端口，不要公开 MySQL / Redis 端口。
- 通过 Nginx / Caddy / SLB 配置 HTTPS 和反向代理。
- 持久化并定期备份 MySQL 数据。
- 记录 `logs/` 并接入进程守护，例如 systemd、supervisor、Docker restart policy 或云平台托管。
