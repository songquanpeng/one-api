package common

import "time"

var StartTime = time.Now().Unix() // unit: second
var Version = "v0.0.0"            // this hard coding will be replaced automatically when building, no need to manually change

const (
	RoleGuestUser  = 0
	RoleCommonUser = 1
	RoleAdminUser  = 10
	RoleRootUser   = 100
)

const (
	UserStatusEnabled  = 1 // don't use 0, 0 is the default value!
	UserStatusDisabled = 2 // also don't use 0
	UserStatusDeleted  = 3
)

const (
	TokenStatusEnabled   = 1 // don't use 0, 0 is the default value!
	TokenStatusDisabled  = 2 // also don't use 0
	TokenStatusExpired   = 3
	TokenStatusExhausted = 4
)

const (
	RedemptionCodeStatusEnabled  = 1 // don't use 0, 0 is the default value!
	RedemptionCodeStatusDisabled = 2 // also don't use 0
	RedemptionCodeStatusUsed     = 3 // also don't use 0
)

const (
	ChannelStatusUnknown          = 0
	ChannelStatusEnabled          = 1 // don't use 0, 0 is the default value!
	ChannelStatusManuallyDisabled = 2 // also don't use 0
	ChannelStatusAutoDisabled     = 3
)

const (
	ChannelTypeUnknown = iota
	ChannelTypeOpenAI
	ChannelTypeAPI2D
	ChannelTypeAzure
	ChannelTypeCloseAI
	ChannelTypeOpenAISB
	ChannelTypeOpenAIMax
	ChannelTypeOhMyGPT
	ChannelTypeCustom
	ChannelTypeAILS
	ChannelTypeAIProxy
	ChannelTypePaLM
	ChannelTypeAPI2GPT
	ChannelTypeAIGC2D
	ChannelTypeAnthropic
	ChannelTypeBaidu
	ChannelTypeZhipu
	ChannelTypeAli
	ChannelTypeXunfei
	ChannelType360
	ChannelTypeOpenRouter
	ChannelTypeAIProxyLibrary
	ChannelTypeFastGPT
	ChannelTypeTencent
	ChannelTypeGemini
	ChannelTypeMoonshot
	ChannelTypeBaichuan
	ChannelTypeMinimax
	ChannelTypeMistral
	ChannelTypeGroq
	ChannelTypeOllama
	ChannelTypeLingYiWanWu

	ChannelTypeDummy
)

var ChannelBaseURLs = []string{
	"",                              // 0
	"https://api.openai.com",        // 1
	"https://oa.api2d.net",          // 2
	"",                              // 3
	"https://api.closeai-proxy.xyz", // 4
	"https://api.openai-sb.com",     // 5
	"https://api.openaimax.com",     // 6
	"https://api.ohmygpt.com",       // 7
	"",                              // 8
	"https://api.caipacity.com",     // 9
	"https://api.aiproxy.io",        // 10
	"https://generativelanguage.googleapis.com", // 11
	"https://api.api2gpt.com",                   // 12
	"https://api.aigc2d.com",                    // 13
	"https://api.anthropic.com",                 // 14
	"https://aip.baidubce.com",                  // 15
	"https://open.bigmodel.cn",                  // 16
	"https://dashscope.aliyuncs.com",            // 17
	"",                                          // 18
	"https://ai.360.cn",                         // 19
	"https://openrouter.ai/api",                 // 20
	"https://api.aiproxy.io",                    // 21
	"https://fastgpt.run/api/openapi",           // 22
	"https://hunyuan.cloud.tencent.com",         // 23
	"https://generativelanguage.googleapis.com", // 24
	"https://api.moonshot.cn",                   // 25
	"https://api.baichuan-ai.com",               // 26
	"https://api.minimax.chat",                  // 27
	"https://api.mistral.ai",                    // 28
	"https://api.groq.com/openai",               // 29
	"http://localhost:11434",                    // 30
	"https://api.lingyiwanwu.com",               // 31
}

const (
	ConfigKeyPrefix = "cfg_"

	ConfigKeyAPIVersion = ConfigKeyPrefix + "api_version"
	ConfigKeyLibraryID  = ConfigKeyPrefix + "library_id"
	ConfigKeyPlugin     = ConfigKeyPrefix + "plugin"
)
