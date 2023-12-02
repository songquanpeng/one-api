package azure

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type AzureProviderFactory struct{}

// 创建 AzureProvider
func (f AzureProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &AzureProvider{
		OpenAIProvider: openai.OpenAIProvider{
			BaseProvider: base.BaseProvider{
				BaseURL:             "",
				Completions:         "/completions",
				ChatCompletions:     "/chat/completions",
				Embeddings:          "/embeddings",
				AudioTranscriptions: "/audio/transcriptions",
				AudioTranslations:   "/audio/translations",
				ImagesGenerations:   "/images/generations",
				// ImagesEdit:          "/images/edit",
				// ImagesVariations:    "/images/variations",
				Context: c,
				// AudioSpeech:         "/audio/speech",
			},
			IsAzure: true,
		},
	}
}

type AzureProvider struct {
	openai.OpenAIProvider
}
