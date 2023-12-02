package aiproxy

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type AIProxyProviderFactory struct{}

func (f AIProxyProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &AIProxyProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.aiproxy.io"),
	}
}

type AIProxyProvider struct {
	*openai.OpenAIProvider
}
