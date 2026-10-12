package internlm

// https://internlm.intern-ai.org.cn/api/document
// 书生（上海人工智能实验室 InternLM）API，OpenAI 兼容：
// Base URL: https://chat.intern-ai.org.cn/api（请求路径 /v1/chat/completions）
// 认证：Authorization: Bearer <API_TOKEN>
// 支持 intern-latest、internlm3、intern-s1（科学多模态）、internvl（视觉）等系列，
// 部分场景免费，具体以官网额度政策为准。

var ModelList = []string{
	"intern-latest",
	"internlm3-latest",
	"intern-s1",
	"intern-s1-pro",
	"internvl2.5-latest",
}
