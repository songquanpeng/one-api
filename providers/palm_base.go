package providers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type PalmProvider struct {
	ProviderConfig
}

// 创建 PalmProvider
func CreatePalmProvider(c *gin.Context) *PalmProvider {
	return &PalmProvider{
		ProviderConfig: ProviderConfig{
			BaseURL:         "https://generativelanguage.googleapis.com",
			ChatCompletions: "/v1beta2/models/chat-bison-001:generateMessage",
			Context:         c,
		},
	}
}

// 获取请求头
func (p *PalmProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)

	headers["Content-Type"] = p.Context.Request.Header.Get("Content-Type")
	headers["Accept"] = p.Context.Request.Header.Get("Accept")
	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json"
	}

	return headers
}

// 获取完整请求 URL
func (p *PalmProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s?key=%s", baseURL, requestURL, p.Context.GetString("api_key"))
}
