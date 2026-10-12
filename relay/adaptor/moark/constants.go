package moark

// https://moark.com/docs
// 模力方舟（Moark，开源中国旗下 AI 平台）Serverless API，OpenAI 兼容：
// Base URL: https://api.moark.com（请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <API_KEY>
// 平台聚合全球主流大模型与国产算力（DeepSeek、千问、智谱、混元、文心、
// 阶跃星辰、书生等），此处仅登记控制台已确认的部分模型，
// 完整列表见 moark.com/serverless-api，其余模型可在渠道中手动添加。

var ModelList = []string{
	"DeepSeek-V4-Pro",
	"DeepSeek-V4-Pro-0813",
	"DeepSeek-V4-Flash",
	"DeepSeek-V4-Flash-0731",
	"DeepSeek-R1-Distill-Qwen-1.5B",
}
