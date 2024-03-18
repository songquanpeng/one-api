package lingyi

import (
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

// 定义供应商工厂
type LingyiProviderFactory struct{}

// 创建 LingyiProvider
// https://platform.lingyiwanwu.com/docs#-create-chat-completion
func (f LingyiProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &LingyiProvider{
		OpenAIProvider: openai.OpenAIProvider{
			BaseProvider: base.BaseProvider{
				Config:    getConfig(),
				Channel:   channel,
				Requester: requester.NewHTTPRequester(*channel.Proxy, openai.RequestErrorHandle),
			},
		},
	}
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.lingyiwanwu.com",
		ChatCompletions: "/v1/chat/completions",
	}
}

type LingyiProvider struct {
	openai.OpenAIProvider
}
