package gemini

import (
	"fmt"
	"one-api/providers/base"
	"strings"

	"github.com/gin-gonic/gin"
)

type GeminiProviderFactory struct{}

// 创建 ClaudeProvider
func (f GeminiProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &GeminiProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://generativelanguage.googleapis.com",
			ChatCompletions: "/",
			Context:         c,
		},
	}
}

type GeminiProvider struct {
	base.BaseProvider
}

func (p *GeminiProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")
	version := "v1"
	if p.Channel.Other != "" {
		version = p.Channel.Other
	}

	return fmt.Sprintf("%s/%s/models/%s:%s", baseURL, version, modelName, requestURL)

}

// 获取请求头
func (p *GeminiProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["x-goog-api-key"] = p.Channel.Key

	return headers
}
