package azure

import (
	"one-api/providers/base"
	"one-api/providers/openai"

	"github.com/gin-gonic/gin"
)

type AzureProvider struct {
	openai.OpenAIProvider
}

// 创建 OpenAIProvider
func CreateAzureProvider(c *gin.Context) *AzureProvider {
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
