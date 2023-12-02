package azureSpeech

import (
	"one-api/providers/base"

	"github.com/gin-gonic/gin"
)

// 定义供应商工厂
type AzureSpeechProviderFactory struct{}

// 创建 AliProvider
func (f AzureSpeechProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &AzureSpeechProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:     "",
			AudioSpeech: "/cognitiveservices/v1",
			Context:     c,
		},
	}
}

type AzureSpeechProvider struct {
	base.BaseProvider
}

// 获取请求头
func (p *AzureSpeechProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	headers["Ocp-Apim-Subscription-Key"] = p.Context.GetString("api_key")
	headers["Content-Type"] = "application/ssml+xml"
	headers["User-Agent"] = "OneAPI"
	// headers["X-Microsoft-OutputFormat"] = "audio-16khz-128kbitrate-mono-mp3"

	return headers
}
