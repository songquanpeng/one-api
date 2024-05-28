package providers

import (
	"one-api/common/config"
	"one-api/model"
	"one-api/providers/ali"
	"one-api/providers/azure"
	azurespeech "one-api/providers/azureSpeech"
	"one-api/providers/baichuan"
	"one-api/providers/baidu"
	"one-api/providers/base"
	"one-api/providers/bedrock"
	"one-api/providers/claude"
	"one-api/providers/cloudflareAI"
	"one-api/providers/cohere"
	"one-api/providers/coze"
	"one-api/providers/deepseek"
	"one-api/providers/gemini"
	"one-api/providers/groq"
	"one-api/providers/hunyuan"
	"one-api/providers/lingyi"
	"one-api/providers/midjourney"
	"one-api/providers/minimax"
	"one-api/providers/mistral"
	"one-api/providers/moonshot"
	"one-api/providers/ollama"
	"one-api/providers/openai"
	"one-api/providers/palm"
	"one-api/providers/stabilityAI"
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
	providerFactories[config.ChannelTypeOpenAI] = openai.OpenAIProviderFactory{}
	providerFactories[config.ChannelTypeAzure] = azure.AzureProviderFactory{}
	providerFactories[config.ChannelTypeAli] = ali.AliProviderFactory{}
	providerFactories[config.ChannelTypeTencent] = tencent.TencentProviderFactory{}
	providerFactories[config.ChannelTypeBaidu] = baidu.BaiduProviderFactory{}
	providerFactories[config.ChannelTypeAnthropic] = claude.ClaudeProviderFactory{}
	providerFactories[config.ChannelTypePaLM] = palm.PalmProviderFactory{}
	providerFactories[config.ChannelTypeZhipu] = zhipu.ZhipuProviderFactory{}
	providerFactories[config.ChannelTypeXunfei] = xunfei.XunfeiProviderFactory{}
	providerFactories[config.ChannelTypeAzureSpeech] = azurespeech.AzureSpeechProviderFactory{}
	providerFactories[config.ChannelTypeGemini] = gemini.GeminiProviderFactory{}
	providerFactories[config.ChannelTypeBaichuan] = baichuan.BaichuanProviderFactory{}
	providerFactories[config.ChannelTypeMiniMax] = minimax.MiniMaxProviderFactory{}
	providerFactories[config.ChannelTypeDeepseek] = deepseek.DeepseekProviderFactory{}
	providerFactories[config.ChannelTypeMistral] = mistral.MistralProviderFactory{}
	providerFactories[config.ChannelTypeGroq] = groq.GroqProviderFactory{}
	providerFactories[config.ChannelTypeBedrock] = bedrock.BedrockProviderFactory{}
	providerFactories[config.ChannelTypeMidjourney] = midjourney.MidjourneyProviderFactory{}
	providerFactories[config.ChannelTypeCloudflareAI] = cloudflareAI.CloudflareAIProviderFactory{}
	providerFactories[config.ChannelTypeCohere] = cohere.CohereProviderFactory{}
	providerFactories[config.ChannelTypeStabilityAI] = stabilityAI.StabilityAIProviderFactory{}
	providerFactories[config.ChannelTypeCoze] = coze.CozeProviderFactory{}
	providerFactories[config.ChannelTypeOllama] = ollama.OllamaProviderFactory{}
	providerFactories[config.ChannelTypeMoonshot] = moonshot.MoonshotProviderFactory{}
	providerFactories[config.ChannelTypeLingyi] = lingyi.LingyiProviderFactory{}
	providerFactories[config.ChannelTypeHunyuan] = hunyuan.HunyuanProviderFactory{}

}

// 获取供应商
func GetProvider(channel *model.Channel, c *gin.Context) base.ProviderInterface {
	factory, ok := providerFactories[channel.Type]
	var provider base.ProviderInterface
	if !ok {
		// 处理未找到的供应商工厂
		baseURL := config.ChannelBaseURLs[channel.Type]
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
