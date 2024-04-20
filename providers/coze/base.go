package coze

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

type CozeProviderFactory struct{}

// 创建 CozeProvider
func (f CozeProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &CozeProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type CozeProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.coze.com/open_api",
		ChatCompletions: "/v2/chat",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	CozeError := &CozeStatus{}
	err := json.NewDecoder(resp.Body).Decode(CozeError)
	if err != nil {
		return nil
	}

	return errorHandle(CozeError)
}

// 错误处理
func errorHandle(CozeError *CozeStatus) *types.OpenAIError {
	if CozeError.Code == 0 {
		return nil
	}
	return &types.OpenAIError{
		Message: CozeError.Msg,
		Type:    "Coze error",
		Code:    CozeError.Code,
	}
}

// 获取请求头
func (p *CozeProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Channel.Key)

	return headers
}

func (p *CozeProvider) GetFullRequestURL(requestURL string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

func convertRole(role string) string {
	switch role {
	case types.ChatMessageRoleSystem, types.ChatMessageRoleAssistant:
		return types.ChatMessageRoleAssistant
	default:
		return types.ChatMessageRoleUser
	}
}
