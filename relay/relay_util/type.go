package relay_util

import (
	"one-api/common/config"
)

var UnknownOwnedBy = "未知"
var ModelOwnedBy map[int]string

func init() {
	ModelOwnedBy = map[int]string{
		config.ChannelTypeOpenAI:       "OpenAI",
		config.ChannelTypeAnthropic:    "Anthropic",
		config.ChannelTypeBaidu:        "Baidu",
		config.ChannelTypePaLM:         "Google PaLM",
		config.ChannelTypeGemini:       "Google Gemini",
		config.ChannelTypeZhipu:        "Zhipu",
		config.ChannelTypeAli:          "Ali",
		config.ChannelTypeXunfei:       "Xunfei",
		config.ChannelType360:          "360",
		config.ChannelTypeTencent:      "Tencent",
		config.ChannelTypeBaichuan:     "Baichuan",
		config.ChannelTypeMiniMax:      "MiniMax",
		config.ChannelTypeDeepseek:     "Deepseek",
		config.ChannelTypeMoonshot:     "Moonshot",
		config.ChannelTypeMistral:      "Mistral",
		config.ChannelTypeGroq:         "Groq",
		config.ChannelTypeLingyi:       "Lingyiwanwu",
		config.ChannelTypeMidjourney:   "Midjourney",
		config.ChannelTypeCloudflareAI: "Cloudflare AI",
		config.ChannelTypeCohere:       "Cohere",
		config.ChannelTypeStabilityAI:  "Stability AI",
		config.ChannelTypeCoze:         "Coze",
		config.ChannelTypeOllama:       "Ollama",
		config.ChannelTypeHunyuan:      "Hunyuan",
	}
}
