package baichuan

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

// 定义供应商工厂
type BaichuanProviderFactory struct{}

// 创建 BaichuanProvider
// https://platform.baichuan-ai.com/docs/api
func (f BaichuanProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &BaichuanProvider{
		OpenAIProvider: openai.OpenAIProvider{
			BaseProvider: base.BaseProvider{
				BaseURL:         "https://api.baichuan-ai.com",
				ChatCompletions: "/v1/chat/completions",
				Embeddings:      "/v1/embeddings",
				Context:         c,
			},
		},
	}
}

type BaichuanProvider struct {
	openai.OpenAIProvider
}
