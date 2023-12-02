package providers

import (
	"one-api/common"
	"one-api/providers/aiproxy"
	"one-api/providers/ali"
	"one-api/providers/api2d"
	"one-api/providers/azure"
	"one-api/providers/baidu"
	"one-api/providers/base"
	"one-api/providers/claude"
	"one-api/providers/closeai"
	"one-api/providers/openai"
	"one-api/providers/openaisb"
	"one-api/providers/palm"
	"one-api/providers/tencent"
	"one-api/providers/xunfei"
	"one-api/providers/zhipu"

	"github.com/gin-gonic/gin"
)

func GetProvider(channelType int, c *gin.Context) base.ProviderInterface {
	switch channelType {
	case common.ChannelTypeOpenAI:
		return openai.CreateOpenAIProvider(c, "")
	case common.ChannelTypeAzure:
		return azure.CreateAzureProvider(c)
	case common.ChannelTypeAli:
		return ali.CreateAliAIProvider(c)
	case common.ChannelTypeTencent:
		return tencent.CreateTencentProvider(c)
	case common.ChannelTypeBaidu:
		return baidu.CreateBaiduProvider(c)
	case common.ChannelTypeAnthropic:
		return claude.CreateClaudeProvider(c)
	case common.ChannelTypePaLM:
		return palm.CreatePalmProvider(c)
	case common.ChannelTypeZhipu:
		return zhipu.CreateZhipuProvider(c)
	case common.ChannelTypeXunfei:
		return xunfei.CreateXunfeiProvider(c)
	case common.ChannelTypeAIProxy:
		return aiproxy.CreateAIProxyProvider(c)
	case common.ChannelTypeAPI2D:
		return api2d.CreateApi2dProvider(c)
	case common.ChannelTypeCloseAI:
		return closeai.CreateCloseaiProxyProvider(c)
	case common.ChannelTypeOpenAISB:
		return openaisb.CreateOpenaiSBProvider(c)
	default:
		baseURL := common.ChannelBaseURLs[channelType]
		if c.GetString("base_url") != "" {
			baseURL = c.GetString("base_url")
		}
		if baseURL != "" {
			return openai.CreateOpenAIProvider(c, baseURL)
		}

		return nil
	}
}
