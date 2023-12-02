package api2gpt

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type Api2gptProviderFactory struct{}

func (f Api2gptProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &Api2gptProvider{
		OpenAIProvider: openai.CreateOpenAIProvider(c, "https://api.api2gpt.com"),
	}
}

type Api2gptProvider struct {
	*openai.OpenAIProvider
}
