package gemini

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

type GeminiProviderFactory struct{}

// 创建 GeminiProvider
func (f GeminiProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &GeminiProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type GeminiProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://generativelanguage.googleapis.com",
		ChatCompletions: "/",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	geminiError := &GeminiErrorResponse{}
	err := json.NewDecoder(resp.Body).Decode(geminiError)
	if err != nil {
		return nil
	}

	return errorHandle(geminiError)
}

// 错误处理
func errorHandle(geminiError *GeminiErrorResponse) *types.OpenAIError {
	if geminiError.Error.Message == "" {
		return nil
	}
	return &types.OpenAIError{
		Message: geminiError.Error.Message,
		Type:    "gemini_error",
		Param:   geminiError.Error.Status,
		Code:    geminiError.Error.Code,
	}
}

func (p *GeminiProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")
	version := "v1beta"
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
