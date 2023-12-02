package palm

import (
	"fmt"
	"one-api/providers/base"
	"strings"

	"github.com/gin-gonic/gin"
)

type PalmProviderFactory struct{}

// 创建 PalmProvider
func (f PalmProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &PalmProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://generativelanguage.googleapis.com",
			ChatCompletions: "/v1beta2/models/chat-bison-001:generateMessage",
			Context:         c,
		},
	}
}

type PalmProvider struct {
	base.BaseProvider
}

// 获取请求头
func (p *PalmProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)

	return headers
}

// 获取完整请求 URL
func (p *PalmProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s?key=%s", baseURL, requestURL, p.Context.GetString("api_key"))
}
