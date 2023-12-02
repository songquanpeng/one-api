package api2d

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type Api2dProviderFactory struct{}

// 创建 Api2dProvider
func (f Api2dProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &Api2dProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://oa.api2d.net"),
	}
}

type Api2dProvider struct {
	*openai.OpenAIProvider
}
