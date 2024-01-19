package ali

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
)

// 定义供应商工厂
type AliProviderFactory struct{}

type AliProvider struct {
	base.BaseProvider
}

// 创建 AliProvider
// https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation
func (f AliProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &AliProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(channel.Proxy, requestErrorHandle),
		},
	}
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://dashscope.aliyuncs.com",
		ChatCompletions: "/api/v1/services/aigc/text-generation/generation",
		Embeddings:      "/api/v1/services/embeddings/text-embedding/text-embedding",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	aliError := &AliError{}
	err := json.NewDecoder(resp.Body).Decode(aliError)
	if err != nil {
		return nil
	}

	return errorHandle(aliError)
}

// 错误处理
func errorHandle(aliError *AliError) *types.OpenAIError {
	if aliError.Code == "" {
		return nil
	}
	return &types.OpenAIError{
		Message: aliError.Message,
		Type:    aliError.Code,
		Param:   aliError.RequestId,
		Code:    aliError.Code,
	}
}

func (p *AliProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	if modelName == "qwen-vl-plus" {
		requestURL = "/api/v1/services/aigc/multimodal-generation/generation"
	}

	return fmt.Sprintf("%s%s", baseURL, requestURL)
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
