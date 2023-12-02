package closeai

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type CloseaiProviderFactory struct{}

// 创建 CloseaiProvider
func (f CloseaiProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &CloseaiProxyProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.closeai-proxy.xyz"),
	}
}

type CloseaiProxyProvider struct {
	*openai.OpenAIProvider
}
