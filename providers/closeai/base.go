package closeai

import (
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type CloseaiProviderFactory struct{}

// 创建 CloseaiProvider
func (f CloseaiProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &CloseaiProxyProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(channel, "https://api.closeai-proxy.xyz"),
	}
}

type CloseaiProxyProvider struct {
	*openai.OpenAIProvider
}
