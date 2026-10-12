package modelscope

// https://www.modelscope.cn/docs/model-service/API-Inference/intro
// 魔搭社区（ModelScope）API-Inference，OpenAI 兼容：
// Base URL: https://api-inference.modelscope.cn（请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <ModelScope Access Token>
// 模型 ID 为魔搭 Model-Id（vendor/model 格式），平台同时提供 Responses API，
// 热门模型有每日免费调用额度，完整支持列表见魔搭模型页。

var ModelList = []string{
	"Qwen/Qwen3.5-35B-A3B",
	"Qwen/Qwen2.5-Coder-32B-Instruct",
	"Qwen/Qwen2.5-VL-72B-Instruct",
	"deepseek-ai/DeepSeek-R1",
	"deepseek-ai/DeepSeek-V4-Pro",
	"ZhipuAI/GLM-5.1",
}
