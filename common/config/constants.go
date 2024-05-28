package config

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

var StartTime = time.Now().Unix() // unit: second
var Version = "v0.0.0"            // this hard coding will be replaced automatically when building, no need to manually change
var SystemName = "One API"
var ServerAddress = "http://localhost:3000"
var Footer = ""
var Logo = ""
var TopUpLink = ""
var ChatLink = ""
var ChatLinks = ""
var QuotaPerUnit = 500 * 1000.0 // $0.002 / 1K tokens
var DisplayInCurrencyEnabled = true
var DisplayTokenStatEnabled = true

// Any options with "Secret", "Token" in its key won't be return by GetOptions

var SessionSecret = uuid.New().String()

var OptionMap map[string]string
var OptionMapRWMutex sync.RWMutex

var ItemsPerPage = 10
var MaxRecentItems = 100

var PasswordLoginEnabled = true
var PasswordRegisterEnabled = true
var EmailVerificationEnabled = false
var GitHubOAuthEnabled = false
var WeChatAuthEnabled = false
var LarkAuthEnabled = false
var TurnstileCheckEnabled = false
var RegisterEnabled = true

// chat cache
var ChatCacheEnabled = false
var ChatCacheExpireMinute = 5 // 5 Minute

// mj
var MjNotifyEnabled = false

var EmailDomainRestrictionEnabled = false
var EmailDomainWhitelist = []string{
	"gmail.com",
	"163.com",
	"126.com",
	"qq.com",
	"outlook.com",
	"hotmail.com",
	"icloud.com",
	"yahoo.com",
	"foxmail.com",
}

var MemoryCacheEnabled = false

var LogConsumeEnabled = true

var SMTPServer = ""
var SMTPPort = 587
var SMTPAccount = ""
var SMTPFrom = ""
var SMTPToken = ""

var ChatImageRequestProxy = ""

var GitHubClientId = ""
var GitHubClientSecret = ""

var LarkClientId = ""
var LarkClientSecret = ""

var WeChatServerAddress = ""
var WeChatServerToken = ""
var WeChatAccountQRCodeImageURL = ""

var TurnstileSiteKey = ""
var TurnstileSecretKey = ""

var QuotaForNewUser = 0
var QuotaForInviter = 0
var QuotaForInvitee = 0
var ChannelDisableThreshold = 5.0
var AutomaticDisableChannelEnabled = false
var AutomaticEnableChannelEnabled = false
var QuotaRemindThreshold = 1000
var PreConsumedQuota = 500
var ApproximateTokenEnabled = false
var DisableTokenEncoders = false
var RetryTimes = 0
var DefaultChannelWeight = uint(1)
var RetryCooldownSeconds = 5

var RootUserEmail = ""

var IsMasterNode = true

var RequestInterval time.Duration

var BatchUpdateEnabled = false
var BatchUpdateInterval = 5

const (
	RoleGuestUser  = 0
	RoleCommonUser = 1
	RoleAdminUser  = 10
	RoleRootUser   = 100
)

var RateLimitKeyExpirationDuration = 20 * time.Minute

const (
	UserStatusEnabled  = 1 // don't use 0, 0 is the default value!
	UserStatusDisabled = 2 // also don't use 0
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
	ChannelTypeUnknown        = 0
	ChannelTypeOpenAI         = 1
	ChannelTypeAPI2D          = 2
	ChannelTypeAzure          = 3
	ChannelTypeCloseAI        = 4
	ChannelTypeOpenAISB       = 5
	ChannelTypeOpenAIMax      = 6
	ChannelTypeOhMyGPT        = 7
	ChannelTypeCustom         = 8
	ChannelTypeAILS           = 9
	ChannelTypeAIProxy        = 10
	ChannelTypePaLM           = 11
	ChannelTypeAPI2GPT        = 12
	ChannelTypeAIGC2D         = 13
	ChannelTypeAnthropic      = 14
	ChannelTypeBaidu          = 15
	ChannelTypeZhipu          = 16
	ChannelTypeAli            = 17
	ChannelTypeXunfei         = 18
	ChannelType360            = 19
	ChannelTypeOpenRouter     = 20
	ChannelTypeAIProxyLibrary = 21
	ChannelTypeFastGPT        = 22
	ChannelTypeTencent        = 23
	ChannelTypeAzureSpeech    = 24
	ChannelTypeGemini         = 25
	ChannelTypeBaichuan       = 26
	ChannelTypeMiniMax        = 27
	ChannelTypeDeepseek       = 28
	ChannelTypeMoonshot       = 29
	ChannelTypeMistral        = 30
	ChannelTypeGroq           = 31
	ChannelTypeBedrock        = 32
	ChannelTypeLingyi         = 33
	ChannelTypeMidjourney     = 34
	ChannelTypeCloudflareAI   = 35
	ChannelTypeCohere         = 36
	ChannelTypeStabilityAI    = 37
	ChannelTypeCoze           = 38
	ChannelTypeOllama         = 39
	ChannelTypeHunyuan        = 40
)

var ChannelBaseURLs = []string{
	"",                                    // 0
	"https://api.openai.com",              // 1
	"https://oa.api2d.net",                // 2
	"",                                    // 3
	"https://api.closeai-proxy.xyz",       // 4
	"https://api.openai-sb.com",           // 5
	"https://api.openaimax.com",           // 6
	"https://api.ohmygpt.com",             // 7
	"",                                    // 8
	"https://api.caipacity.com",           // 9
	"https://api.aiproxy.io",              // 10
	"",                                    // 11
	"https://api.api2gpt.com",             // 12
	"https://api.aigc2d.com",              // 13
	"https://api.anthropic.com",           // 14
	"https://aip.baidubce.com",            // 15
	"https://open.bigmodel.cn",            // 16
	"https://dashscope.aliyuncs.com",      // 17
	"",                                    // 18
	"https://ai.360.cn",                   // 19
	"https://openrouter.ai/api",           // 20
	"https://api.aiproxy.io",              // 21
	"https://fastgpt.run/api/openapi",     // 22
	"https://hunyuan.cloud.tencent.com",   //23
	"",                                    //24
	"",                                    //25
	"https://api.baichuan-ai.com",         //26
	"https://api.minimax.chat/v1",         //27
	"https://api.deepseek.com",            //28
	"https://api.moonshot.cn",             //29
	"https://api.mistral.ai",              //30
	"https://api.groq.com/openai",         //31
	"",                                    //32
	"https://api.lingyiwanwu.com",         //33
	"",                                    //34
	"",                                    //35
	"https://api.cohere.ai/v1",            //36
	"https://api.stability.ai/v2beta",     //37
	"https://api.coze.com/open_api",       //38
	"",                                    //39
	"https://hunyuan.tencentcloudapi.com", //40
}

const (
	RelayModeUnknown = iota
	RelayModeChatCompletions
	RelayModeCompletions
	RelayModeEmbeddings
	RelayModeModerations
	RelayModeImagesGenerations
	RelayModeImagesEdits
	RelayModeImagesVariations
	RelayModeEdits
	RelayModeAudioSpeech
	RelayModeAudioTranscription
	RelayModeAudioTranslation
)
