package openaisb

import (
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type OpenaiSBProvider struct {
	*openai.OpenAIProvider
}

// 创建 OpenaiSBProvider
func CreateOpenaiSBProvider(c *gin.Context) *OpenaiSBProvider {
	return &OpenaiSBProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.openai-sb.com"),
	}
}
