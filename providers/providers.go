package providers

import (
	"one-api/common"
	"one-api/model"
	"one-api/providers/aigc2d"
	"one-api/providers/aiproxy"
	"one-api/providers/ali"
	"one-api/providers/api2d"
	"one-api/providers/api2gpt"
	"one-api/providers/azure"
	azurespeech "one-api/providers/azureSpeech"
	"one-api/providers/baichuan"
	"one-api/providers/baidu"
	"one-api/providers/base"
	"one-api/providers/bedrock"
	"one-api/providers/claude"
	"one-api/providers/closeai"
	"one-api/providers/deepseek"
	"one-api/providers/gemini"
	"one-api/providers/groq"
	"one-api/providers/minimax"
	"one-api/providers/mistral"
	"one-api/providers/openai"
	"one-api/providers/openaisb"
	"one-api/providers/palm"
	"one-api/providers/tencent"
	"one-api/providers/xunfei"
	"one-api/providers/zhipu"

	"github.com/gin-gonic/gin"
)

// 定义供应商工厂接口
type ProviderFactory interface {
	Create(Channel *model.Channel) base.ProviderInterface
}

// 创建全局的供应商工厂映射
var providerFactories = make(map[int]ProviderFactory)

// 在程序启动时，添加所有的供应商工厂
func init() {
	providerFactories[common.ChannelTypeOpenAI] = openai.OpenAIProviderFactory{}
	providerFactories[common.ChannelTypeAzure] = azure.AzureProviderFactory{}
	providerFactories[common.ChannelTypeAli] = ali.AliProviderFactory{}
	providerFactories[common.ChannelTypeTencent] = tencent.TencentProviderFactory{}
	providerFactories[common.ChannelTypeBaidu] = baidu.BaiduProviderFactory{}
	providerFactories[common.ChannelTypeAnthropic] = claude.ClaudeProviderFactory{}
	providerFactories[common.ChannelTypePaLM] = palm.PalmProviderFactory{}
	providerFactories[common.ChannelTypeZhipu] = zhipu.ZhipuProviderFactory{}
	providerFactories[common.ChannelTypeXunfei] = xunfei.XunfeiProviderFactory{}
	providerFactories[common.ChannelTypeAIProxy] = aiproxy.AIProxyProviderFactory{}
	providerFactories[common.ChannelTypeAPI2D] = api2d.Api2dProviderFactory{}
	providerFactories[common.ChannelTypeCloseAI] = closeai.CloseaiProviderFactory{}
	providerFactories[common.ChannelTypeOpenAISB] = openaisb.OpenaiSBProviderFactory{}
	providerFactories[common.ChannelTypeAIGC2D] = aigc2d.Aigc2dProviderFactory{}
	providerFactories[common.ChannelTypeAPI2GPT] = api2gpt.Api2gptProviderFactory{}
	providerFactories[common.ChannelTypeAzureSpeech] = azurespeech.AzureSpeechProviderFactory{}
	providerFactories[common.ChannelTypeGemini] = gemini.GeminiProviderFactory{}
	providerFactories[common.ChannelTypeBaichuan] = baichuan.BaichuanProviderFactory{}
	providerFactories[common.ChannelTypeMiniMax] = minimax.MiniMaxProviderFactory{}
	providerFactories[common.ChannelTypeDeepseek] = deepseek.DeepseekProviderFactory{}
	providerFactories[common.ChannelTypeMistral] = mistral.MistralProviderFactory{}
	providerFactories[common.ChannelTypeGroq] = groq.GroqProviderFactory{}
	providerFactories[common.ChannelTypeBedrock] = bedrock.BedrockProviderFactory{}

}

// 获取供应商
func GetProvider(channel *model.Channel, c *gin.Context) base.ProviderInterface {
	factory, ok := providerFactories[channel.Type]
	var provider base.ProviderInterface
	if !ok {
		// 处理未找到的供应商工厂
		baseURL := common.ChannelBaseURLs[channel.Type]
		if channel.GetBaseURL() != "" {
			baseURL = channel.GetBaseURL()
		}
		if baseURL == "" {
			return nil
		}

		provider = openai.CreateOpenAIProvider(channel, baseURL)
	} else {
		provider = factory.Create(channel)
	}
	provider.SetContext(c)

	return provider
}
