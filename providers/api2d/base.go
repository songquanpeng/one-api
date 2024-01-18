package api2d

import (
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type Api2dProviderFactory struct{}

// 创建 Api2dProvider
func (f Api2dProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &Api2dProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(channel, "https://oa.api2d.net"),
	}
}

type Api2dProvider struct {
	*openai.OpenAIProvider
}
