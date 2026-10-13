package amdtokenfactory

// https://developer.amd.com.cn/radeon/tokenfactory
// AMD Radeon Cloud Token Factory，OpenAI 兼容：
// Base URL: https://developer.amd.com.cn/radeon/api（请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <API_KEY>（同一密钥兼容 Anthropic Messages 协议）
// Public Free Model APIs 免费开放（每日有额度上限），此处登记的模型当前免费，
// 如需计费请在系统设置中按需调整模型倍率。

var ModelList = []string{
	"DeepSeek-V4-Flash",
	"DeepSeek-V4-Flash-Vision-Exp",
	"Qwen3.8-Flash-Next",
	"MiniCPM5-1B",
	"GLM-5.3-Flash",
}
