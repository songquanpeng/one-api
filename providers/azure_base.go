package providers

import (
	"github.com/gin-gonic/gin"
)

type AzureProvider struct {
	OpenAIProvider
}

// 创建 OpenAIProvider
func CreateAzureProvider(c *gin.Context) *AzureProvider {
	return &AzureProvider{
		OpenAIProvider: OpenAIProvider{
			ProviderConfig: ProviderConfig{
				BaseURL:             "",
				Completions:         "/completions",
				ChatCompletions:     "/chat/completions",
				Embeddings:          "/embeddings",
				AudioSpeech:         "/audio/speech",
				AudioTranscriptions: "/audio/transcriptions",
				AudioTranslations:   "/audio/translations",
				Context:             c,
			},
			isAzure: true,
		},
	}
}

// // 获取完整请求 URL
// func (p *AzureProvider) GetFullRequestURL(requestURL string, modelName string) string {
// 	apiVersion := p.Context.GetString("api_version")
// 	requestURL = fmt.Sprintf("/openai/deployments/%s/%s?api-version=%s", modelName, requestURL, apiVersion)
// 	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

// 	if strings.HasPrefix(baseURL, "https://gateway.ai.cloudflare.com") {
// 		requestURL = strings.TrimPrefix(requestURL, "/openai/deployments")
// 	}

// 	return fmt.Sprintf("%s%s", baseURL, requestURL)
// }
