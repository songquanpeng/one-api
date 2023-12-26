package ali

import (
	"fmt"

	"one-api/providers/base"

	"github.com/gin-gonic/gin"
)

// 定义供应商工厂
type AliProviderFactory struct{}

// 创建 AliProvider
// https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
func (f AliProviderFactory) Create(c *gin.Context) base.ProviderInterface {
	return &AliProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://dashscope.aliyuncs.com",
			ChatCompletions: "/api/v1/services/aigc/text-generation/generation",
			Embeddings:      "/api/v1/services/embeddings/text-embedding/text-embedding",
			Context:         c,
		},
	}
}

type AliProvider struct {
	base.BaseProvider
}

// 获取请求头
func (p *AliProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Channel.Key)
	if p.Channel.Other != "" {
		headers["X-DashScope-Plugin"] = p.Channel.Other
	}

	return headers
}
