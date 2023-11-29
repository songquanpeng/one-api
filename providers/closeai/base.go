package closeai

import (
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type CloseaiProxyProvider struct {
	*openai.OpenAIProvider
}

// 创建 CloseaiProxyProvider
func CreateCloseaiProxyProvider(c *gin.Context) *CloseaiProxyProvider {
	return &CloseaiProxyProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.closeai-proxy.xyz"),
	}
}
