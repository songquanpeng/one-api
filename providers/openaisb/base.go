package openaisb

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type OpenaiSBProviderFactory struct{}

// 创建 OpenaiSBProvider
func (f OpenaiSBProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &OpenaiSBProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.openai-sb.com"),
	}
}

type OpenaiSBProvider struct {
	*openai.OpenAIProvider
}
