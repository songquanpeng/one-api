package stabilityAI

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

type StabilityAIProviderFactory struct{}

// 创建 StabilityAIProvider
func (f StabilityAIProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &StabilityAIProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type StabilityAIProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:           "https://api.stability.ai/v2beta",
		ImagesGenerations: "/stable-image/generate",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	stabilityAIError := &StabilityAIError{}
	err := json.NewDecoder(resp.Body).Decode(stabilityAIError)
	if err != nil {
		return nil
	}

	return errorHandle(stabilityAIError)
}

// 错误处理
func errorHandle(stabilityAIError *StabilityAIError) *types.OpenAIError {
	openaiError := &types.OpenAIError{
		Type: "stabilityAI_error",
	}

	if stabilityAIError.Name != "" {
		openaiError.Message = stabilityAIError.String()
		openaiError.Code = stabilityAIError.Name
	} else {
		openaiError.Message = stabilityAIError.Message
		openaiError.Code = "stabilityAI_error"
	}

	return openaiError
}

func (p *StabilityAIProvider) GetFullRequestURL(requestURL string, modelName string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s/%s", baseURL, requestURL, modelName)
}

// 获取请求头
func (p *StabilityAIProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = "Bearer " + p.Channel.Key

	return headers
}
