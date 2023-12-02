package aiproxy

import (
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type AIProxyProvider struct {
	*openai.OpenAIProvider
}

// 创建 CreateAIProxyProvider
func CreateAIProxyProvider(c *gin.Context) *AIProxyProvider {
	return &AIProxyProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.aiproxy.io"),
	}
}
