package lmstudio

// https://lmstudio.ai/docs/developer/openai-compat
// LM Studio 本地推理服务，OpenAI 兼容：
// Base URL: http://localhost:1234（请求路径 /v1/chat/completions）
// 认证：任意非空 API Key（如 lm-studio）
// 模型为本地已加载的模型（GET /v1/models 仅返回已加载模型），
// 此处仅登记常见开源模型作为默认值，请按本地实际加载的模型增删。

var ModelList = []string{
	"openai/gpt-oss-20b",
	"openai/gpt-oss-120b",
	"qwen/qwen3-coder-30b",
	"qwen/qwen3-32b",
	"google/gemma-3-27b-it",
	"meta-llama/llama-3.3-70b-instruct",
}
