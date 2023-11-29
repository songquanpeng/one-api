package ali

import (
	"fmt"

	"one-api/providers/base"

	"github.com/gin-gonic/gin"
)

type AliProvider struct {
	base.BaseProvider
}

// https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
// 创建 AliAIProvider
func CreateAliAIProvider(c *gin.Context) *AliProvider {
	return &AliProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://dashscope.aliyuncs.com",
			ChatCompletions: "/api/v1/services/aigc/text-generation/generation",
			Embeddings:      "/api/v1/services/embeddings/text-embedding/text-embedding",
			Context:         c,
		},
	}
}

// 获取请求头
func (p *AliProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Context.GetString("api_key"))

	return headers
}
