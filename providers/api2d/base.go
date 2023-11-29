package api2d

import (
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type Api2dProvider struct {
	*openai.OpenAIProvider
}

// 创建 Api2dProvider
func CreateApi2dProvider(c *gin.Context) *Api2dProvider {
	return &Api2dProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://oa.api2d.net"),
	}
}
