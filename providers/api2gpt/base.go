package api2gpt

import (
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type Api2gptProviderFactory struct{}

func (f Api2gptProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &Api2gptProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(channel, "https://api.api2gpt.com"),
	}
}

type Api2gptProvider struct {
	*openai.OpenAIProvider
}
