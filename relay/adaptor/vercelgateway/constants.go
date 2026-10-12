package vercelgateway

// https://vercel.com/docs/ai-gateway
// Vercel AI Gateway 为统一模型网关，OpenAI Chat Completions 兼容：
// Base URL: https://ai-gateway.vercel.sh（请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <AI_GATEWAY_API_KEY>（或 Vercel OIDC Token）
// 模型 ID 为 "provider/model" 格式，网关覆盖 200+ 模型并按上游价格透传计费，
// 此处仅登记常用模型，其余模型可在渠道中手动添加（支持 GET /v1/models 查询）。

var ModelList = []string{
	// OpenAI
	"openai/gpt-5",
	"openai/gpt-5-mini",
	"openai/gpt-5-nano",
	"openai/gpt-4.1",
	"openai/o3",
	// Anthropic
	"anthropic/claude-sonnet-5",
	"anthropic/claude-opus-5",
	"anthropic/claude-sonnet-4.5",
	"anthropic/claude-haiku-4.5",
	// Google
	"google/gemini-2.5-pro",
	"google/gemini-2.5-flash",
	// xAI
	"xai/grok-4",
	// DeepSeek
	"deepseek/deepseek-r1",
	// Meta
	"meta/llama-4-maverick",
	// Mistral
	"mistral/mistral-large",
}
