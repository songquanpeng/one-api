package deepseek

import (
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type DeepseekProviderFactory struct{}

// 创建 DeepseekProvider
func (f DeepseekProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	config := getDeepseekConfig()
	return &DeepseekProvider{
		OpenAIProvider: openai.OpenAIProvider{
			BaseProvider: base.BaseProvider{
				Config:    config,
				Channel:   channel,
				Requester: requester.NewHTTPRequester(*channel.Proxy, openai.RequestErrorHandle),
			},
			BalanceAction: false,
		},
	}
}

func getDeepseekConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.deepseek.com",
		ChatCompletions: "/v1/chat/completions",
		ModelList:       "/v1/models",
	}
}

type DeepseekProvider struct {
	openai.OpenAIProvider
}
