package palm

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

type PalmProviderFactory struct{}

// 创建 PalmProvider
func (f PalmProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &PalmProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type PalmProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://generativelanguage.googleapis.com",
		ChatCompletions: "/v1beta2/models/chat-bison-001:generateMessage",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	palmError := &PaLMErrorResponse{}
	err := json.NewDecoder(resp.Body).Decode(palmError)
	if err != nil {
		return nil
	}

	return errorHandle(palmError)
}

// 错误处理
func errorHandle(palmError *PaLMErrorResponse) *types.OpenAIError {
	if palmError.Error.Code == 0 {
		return nil
	}
	return &types.OpenAIError{
		Message: palmError.Error.Message,
		Type:    "palm_error",
		Param:   palmError.Error.Status,
		Code:    palmError.Error.Code,
	}
}

// 获取请求头
func (p *PalmProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["x-goog-api-key"] = p.Channel.Key

	return headers
}

// 获取完整请求 URL
func (p *PalmProvider) GetFullRequestURL(requestURL string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}
