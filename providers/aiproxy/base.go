package aiproxy

import (
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type AIProxyProviderFactory struct{}

type AIProxyProvider struct {
	*openai.OpenAIProvider
}

func (f AIProxyProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &AIProxyProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(channel, "https://api.aiproxy.io"),
	}
}
