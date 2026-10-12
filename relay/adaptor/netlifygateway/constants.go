package netlifygateway

// https://docs.netlify.com/build/ai-gateway/overview/
// Netlify AI Gateway 为统一模型网关，OpenAI Chat Completions 兼容：
// Base URL: https://ai-gateway.netlify.com（OpenAI SDK 使用 <base>/v1，请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <AI_GATEWAY_API_KEY>（Netlify 注入的 NETLIFY_AI_GATEWAY_KEY）
// OpenAI / Anthropic / Gemini 模型由网关直接托管，模型 ID 与官方一致；
// 其余模型（xAI、DeepSeek、Meta、Mistral、Qwen 等）经 OpenRouter 路由，
// 模型 ID 采用 OpenRouter 记法（如 "deepseek/deepseek-chat"）。
// 计费按上游官方价透传（Netlify credits），此处仅登记常用模型，
// 其余模型可在渠道中手动添加（支持每模型独立定价调整）。

var ModelList = []string{
	// OpenAI（网关直接托管）
	"gpt-5",
	"gpt-5-mini",
	"gpt-5-nano",
	"gpt-5-pro",
	"gpt-4.1",
	"gpt-4.1-mini",
	"gpt-4.1-nano",
	"gpt-4o",
	"gpt-4o-mini",
	"o3",
	"o3-mini",
	"o4-mini",
	// Anthropic（网关直接托管）
	"claude-opus-4-5",
	"claude-sonnet-4-5",
	"claude-haiku-4-5",
	// Google Gemini（网关直接托管）
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
	// 经 OpenRouter 路由（OpenRouter 记法）
	"deepseek/deepseek-chat",
	"deepseek/deepseek-r1",
	"moonshotai/kimi-k2",
	"qwen/qwen3-235b-a22b-2507",
	"meta-llama/llama-4-maverick",
	"mistralai/mistral-large-4-0",
	"x-ai/grok-4.7",
	"z-ai/glm-5.3",
	"perplexity/sonar-pro",
}
