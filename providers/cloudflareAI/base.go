package cloudflareAI

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/requester"
	"one-api/model"
	"one-api/providers/base"
	"one-api/types"
	"strings"
)

type CloudflareAIProviderFactory struct{}

// 创建 CloudflareAIProvider
func (f CloudflareAIProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	cf := &CloudflareAIProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}

	tokens := strings.Split(channel.Key, "|")
	if len(tokens) == 2 {
		cf.AccountID = tokens[0]
		cf.CFToken = tokens[1]
	}

	return cf
}

type CloudflareAIProvider struct {
	base.BaseProvider
	AccountID string
	CFToken   string
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:             "https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s",
		ImagesGenerations:   "true",
		ChatCompletions:     "true",
		AudioTranscriptions: "true",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	CloudflareAIError := &CloudflareAIError{}
	err := json.NewDecoder(resp.Body).Decode(CloudflareAIError)
	if err != nil {
		return nil
	}

	return errorHandle(CloudflareAIError)
}

// 错误处理
func errorHandle(CloudflareAIError *CloudflareAIError) *types.OpenAIError {
	if CloudflareAIError.Success || len(CloudflareAIError.Error) == 0 {
		return nil
	}
	return &types.OpenAIError{
		Message: CloudflareAIError.Error[0].Message,
		Type:    "CloudflareAI error",
		Code:    CloudflareAIError.Error[0].Code,
	}
}

// 获取请求头
func (p *CloudflareAIProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.CFToken)

	return headers
}

func (p *CloudflareAIProvider) GetFullRequestURL(modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf(baseURL, p.AccountID, modelName)
}
