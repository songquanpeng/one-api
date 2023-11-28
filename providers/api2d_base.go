package providers

import "github.com/gin-gonic/gin"

type Api2dProvider struct {
	*OpenAIProvider
}

// 创建 OpenAIProvider
func CreateApi2dProvider(c *gin.Context) *Api2dProvider {
	return &Api2dProvider{
		OpenAIProvider: CreateOpenAIProvider(c, "https://oa.api2d.net"),
	}
}
