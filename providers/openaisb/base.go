package openaisb

import (
	"one-api/model"
	"one-api/providers/base"
	"one-api/providers/openai"
)

type OpenaiSBProviderFactory struct{}

// 创建 OpenaiSBProvider
func (f OpenaiSBProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &OpenaiSBProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(channel, "https://api.openai-sb.com"),
	}
}

type OpenaiSBProvider struct {
	*openai.OpenAIProvider
}
