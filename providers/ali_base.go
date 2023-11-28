package providers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type AliAIProvider struct {
	ProviderConfig
}

type AliError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type AliUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// 创建 AliAIProvider
// https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
func CreateAliAIProvider(c *gin.Context) *AliAIProvider {
	return &AliAIProvider{
		ProviderConfig: ProviderConfig{
			BaseURL:         "https://dashscope.aliyuncs.com",
			ChatCompletions: "/api/v1/services/aigc/text-generation/generation",
			Embeddings:      "/api/v1/services/embeddings/text-embedding/text-embedding",
			Context:         c,
		},
	}
}

// 获取请求头
func (p *AliAIProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Context.GetString("api_key"))

	headers["Content-Type"] = p.Context.Request.Header.Get("Content-Type")
	headers["Accept"] = p.Context.Request.Header.Get("Accept")
	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json"
	}

	return headers
}
