package aigc2d

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type Aigc2dProviderFactory struct{}

func (f Aigc2dProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &Aigc2dProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.aigc2d.com"),
	}
}

type Aigc2dProvider struct {
	*openai.OpenAIProvider
}
