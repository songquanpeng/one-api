package aigc2d

import (
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type Aigc2dProviderFactory struct{}

type Aigc2dProvider struct {
	*openai.OpenAIProvider
}

func (f Aigc2dProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &Aigc2dProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(channel, "https://api.aigc2d.com"),
	}
}
