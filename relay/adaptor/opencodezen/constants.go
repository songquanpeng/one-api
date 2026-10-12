package opencodezen

// https://opencode.ai/docs/zen/
// OpenCode Zen 为按量付费的模型网关，不同模型走不同协议端点：
// - /v1/chat/completions（OpenAI 兼容）：GLM、Kimi、Qwen3 Coder、MiniMax 及免费档模型
// - /v1/responses（OpenAI Responses）：GPT 系列（建议直接使用 OpenAI 渠道或本网关 /v1/responses）
// - /v1/messages（Anthropic 兼容）：Claude 系列
// - /v1beta/models（Google 兼容）：Gemini 系列
// 此处仅登记走 chat/completions 端点的模型，其余模型可在渠道中手动添加。

var ModelList = []string{
	// OpenAI 兼容端点（chat/completions）
	"glm-4.6",
	"glm-4.7-free",
	"kimi-k2",
	"kimi-k2-thinking",
	"qwen3-coder-480b",
	"minimax-m2.1-free",
	// 限时免费档（$0.00/1M tokens，数据可能被用于训练，勿传敏感数据）
	"big-pickle",
	"mimo-v2.5-flash-free",
	"ling-3.0-flash-fin-free",
	"nemotron-3-ultra-free",
	"nemotron-3.5-lightning-free",
}
