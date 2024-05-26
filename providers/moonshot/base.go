package moonshot

import (
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type MoonshotProviderFactory struct{}

// 创建 MoonshotProvider
func (f MoonshotProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	config := getMoonshotConfig()
	return &MoonshotProvider{
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

func getMoonshotConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.moonshot.cn",
		ChatCompletions: "/v1/chat/completions",
		ModelList:       "/v1/models",
	}
}

type MoonshotProvider struct {
	openai.OpenAIProvider
}
