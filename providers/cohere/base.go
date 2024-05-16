package cohere

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

type CohereProviderFactory struct{}

// 创建 CohereProvider
func (f CohereProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &CohereProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type CohereProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.cohere.ai/v1",
		ChatCompletions: "/chat",
		ModelList:       "/models",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	CohereError := &CohereError{}
	err := json.NewDecoder(resp.Body).Decode(CohereError)
	if err != nil {
		return nil
	}

	return errorHandle(CohereError)
}

// 错误处理
func errorHandle(CohereError *CohereError) *types.OpenAIError {
	if CohereError.Message == "" {
		return nil
	}
	return &types.OpenAIError{
		Message: CohereError.Message,
		Type:    "Cohere error",
	}
}

// 获取请求头
func (p *CohereProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", p.Channel.Key)

	return headers
}

func (p *CohereProvider) GetFullRequestURL(requestURL string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")

	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

func convertRole(role string) string {
	switch role {
	case types.ChatMessageRoleSystem:
		return "SYSTEM"
	case types.ChatMessageRoleAssistant:
		return "CHATBOT"
	default:
		return "USER"
	}
}
