package baichuan

import (
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

// 定义供应商工厂
type BaichuanProviderFactory struct{}

// 创建 BaichuanProvider
// https://platform.baichuan-ai.com/docs/api
func (f BaichuanProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &BaichuanProvider{
		OpenAIProvider: openai.OpenAIProvider{
			BaseProvider: base.BaseProvider{
				Config:    getConfig(),
				Channel:   channel,
				Requester: requester.NewHTTPRequester(channel.Proxy, openai.RequestErrorHandle),
			},
		},
	}
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.baichuan-ai.com",
		ChatCompletions: "/v1/chat/completions",
		Embeddings:      "/v1/embeddings",
	}
}

type BaichuanProvider struct {
	openai.OpenAIProvider
}
