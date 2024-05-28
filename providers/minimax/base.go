package minimax

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

type MiniMaxProviderFactory struct{}

// 创建 MiniMaxProvider
func (f MiniMaxProviderFactory) Create(channel *model.Channel) base.ProviderInterface {
	return &MiniMaxProvider{
		BaseProvider: base.BaseProvider{
			Config:    getConfig(),
			Channel:   channel,
			Requester: requester.NewHTTPRequester(*channel.Proxy, requestErrorHandle),
		},
	}
}

type MiniMaxProvider struct {
	base.BaseProvider
}

func getConfig() base.ProviderConfig {
	return base.ProviderConfig{
		BaseURL:         "https://api.minimax.chat/v1",
		ChatCompletions: "/text/chatcompletion_pro",
		Embeddings:      "/embeddings",
	}
}

// 请求错误处理
func requestErrorHandle(resp *http.Response) *types.OpenAIError {
	minimaxError := &MiniMaxBaseResp{}
	err := json.NewDecoder(resp.Body).Decode(minimaxError)
	if err != nil {
		return nil
	}

	return errorHandle(&minimaxError.BaseResp)
}

// 错误处理
func errorHandle(minimaxError *BaseResp) *types.OpenAIError {
	if minimaxError.StatusCode == 0 {
		return nil
	}
	return &types.OpenAIError{
		Message: minimaxError.StatusMsg,
		Type:    "minimax_error",
		Code:    minimaxError.StatusCode,
	}
}

func (p *MiniMaxProvider) GetFullRequestURL(requestURL string) string {
	baseURL := strings.TrimSuffix(p.GetBaseURL(), "/")
	keys := strings.Split(p.Channel.Key, "|")
	if len(keys) != 2 {
		return ""
	}

	return fmt.Sprintf("%s%s?GroupId=%s", baseURL, requestURL, keys[1])
}

// 获取请求头
func (p *MiniMaxProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)
	keys := strings.Split(p.Channel.Key, "|")

	headers["Authorization"] = "Bearer " + keys[0]

	return headers
}

func defaultBot() MiniMaxBotSetting {
	return MiniMaxBotSetting{
		BotName: types.ChatMessageRoleAssistant,
		Content: "You are a helpful assistant. You can help me by answering my questions. You can also ask me questions.",
	}
}

func defaultReplyConstraints() ReplyConstraints {
	return ReplyConstraints{
		SenderType: "BOT",
		SenderName: types.ChatMessageRoleAssistant,
	}
}

func convertRole(roleName string) (string, string) {
	switch roleName {
	case types.ChatMessageRoleTool, types.ChatMessageRoleFunction:
		return "FUNCTION", types.ChatMessageRoleAssistant
	case types.ChatMessageRoleSystem, types.ChatMessageRoleAssistant:
		return "BOT", types.ChatMessageRoleAssistant
	default:
		return "USER", types.ChatMessageRoleUser
	}
}

func convertFinishReason(finishReason string) string {
	switch finishReason {
	case "max_output":
		return types.FinishReasonLength
	default:
		return finishReason
	}
}
