# Cloudflare Containers 部署模板（可选路径）

仓库已预置 Cloudflare Containers 全套配置（2026-04 GA，需 Workers Paid $5/月）：

| 文件 | 用途 |
|------|------|
| `wrangler.jsonc`（仓库根目录） | Worker + 容器定义（0.5 vCPU / 1GB / 4GB 磁盘，单实例） |
| `worker/` | 路由 Worker（Durable Object 封装容器，空闲 30 分钟休眠） |
| `deploy-cloudflare.yml.disabled` | GitHub Actions 部署工作流模板 |

## 启用步骤
1. 确认 Cloudflare 账号已开通 Workers Paid
2. `mv docs/deploy-templates/deploy-cloudflare.yml.disabled .github/workflows/deploy-cloudflare.yml`
3. GitHub 仓库 Settings → Secrets → 添加 `CLOUDFLARE_API_TOKEN`
   （权限: Workers Scripts:Edit + Containers:Edit + Durable Objects:Edit）
4. 设置 secrets（本地或 Actions 内执行）:
   - `npx wrangler secret put SQL_DSN`（可留空 = 容器内 SQLite）
   - `npx wrangler secret put SESSION_SECRET`
5. push 到 main 即自动构建部署（Actions runner 自带 Docker）

数据库当前策略: 容器内 SQLite（`/data/one-api.db`），容器实例重启数据重置；
要持久化随时可在 secrets 加 `SQL_DSN` 切到外部 Postgres/MySQL，代码零改动。
