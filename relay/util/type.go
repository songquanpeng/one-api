package util

import "one-api/common"

var UnknownOwnedBy = "未知"
var ModelOwnedBy map[int]string

func init() {
	ModelOwnedBy = map[int]string{
		common.ChannelTypeOpenAI:    "OpenAI",
		common.ChannelTypeAnthropic: "Anthropic",
		common.ChannelTypeBaidu:     "Baidu",
		common.ChannelTypePaLM:      "Google PaLM",
		common.ChannelTypeGemini:    "Google Gemini",
		common.ChannelTypeZhipu:     "Zhipu",
		common.ChannelTypeAli:       "Ali",
		common.ChannelTypeXunfei:    "Xunfei",
		common.ChannelType360:       "360",
		common.ChannelTypeTencent:   "Tencent",
		common.ChannelTypeBaichuan:  "Baichuan",
		common.ChannelTypeMiniMax:   "MiniMax",
		common.ChannelTypeDeepseek:  "Deepseek",
		common.ChannelTypeMoonshot:  "Moonshot",
		common.ChannelTypeMistral:   "Mistral",
		common.ChannelTypeGroq:      "Groq",
		common.ChannelTypeLingyi:    "Lingyiwanwu",
	}
}
