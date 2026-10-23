import { Container, getContainer } from "@cloudflare/containers";

/**
 * One API 容器封装
 * - 镜像: 仓库根目录 Dockerfile 多阶段构建（前端三主题 + Go 二进制）
 * - defaultPort 必须与容器内 one-api 监听端口一致（PORT 环境变量）
 * - 所有请求路由到固定 ID "oneapi" 的单一实例，保证 SQLite/会话状态一致
 *   （如外接 Supabase Postgres，可改为多实例；当前单实例最省成本）
 */
export class OneAPIContainer extends Container {
	defaultPort = 3000;
	// LLM 流式转发场景，空闲 30 分钟后休眠（按活跃 CPU 计费，休眠不花钱）
	sleepAfter = "30m";

	constructor(ctx: DurableObjectState, env: Env) {
		super(ctx, env);
		// Worker Secrets（wrangler secret put 注入）透传给容器进程
		this.envVars = {
			PORT: "3000",
			TZ: "Asia/Shanghai",
			SQL_DSN: env.SQL_DSN ?? "",
			SESSION_SECRET: env.SESSION_SECRET ?? "change-me-random-string",
			STREAMING_TIMEOUT: "300",
		};
	}
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		// 固定 ID → 全部请求进入同一个容器实例
		return getContainer(env.ONEAPI_CONTAINER, "oneapi").fetch(request);
	},
};
