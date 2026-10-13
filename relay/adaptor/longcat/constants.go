package longcat

// https://longcat.chat/platform/docs/zh/
// 美团 LongCat（龙猫）API 开放平台，OpenAI 兼容：
// Base URL: https://api.longcat.chat/openai（请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <API_KEY>
// 模型 ID 大小写敏感，LongCat-2.0 支持 1M 上下文与 128K 输出。

var ModelList = []string{
	"LongCat-2.5-Preview",
	"LongCat-2.0",
	"LongCat-Flash-Chat",
	"LongCat-Flash-Lite",
}
