package azureSpeech

import (
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
)

// 定义供应商工厂
type AzureSpeechProviderFactory struct{}

// 创建 AliProvider
func (f AzureSpeechProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &AzureSpeechProvider{
		BaseProvider: base.BaseProvider{
			Config: base.ProviderConfig{
				AudioSpeech: "/cognitiveservices/v1",
			},
			Channel:   channel,
			Requester: requester.NewHTTPRequester(channel.Proxy, nil),
		},
	}
}

type AzureSpeechProvider struct {
	base.BaseProvider
}

// 获取请求头
func (p *AzureSpeechProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	headers["Ocp-Apim-Subscription-Key"] = p.Channel.Key
	headers["Content-Type"] = "application/ssml+xml"
	headers["User-Agent"] = "OneAPI"
	// headers["X-Microsoft-OutputFormat"] = "audio-16khz-128kbitrate-mono-mp3"

	return headers
}
