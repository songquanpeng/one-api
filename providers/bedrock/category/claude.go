package category

import (
	"encoding/json"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/providers/base"
	"one-api/providers/claude"
	"one-api/types"
)

const anthropicVersion = "bedrock-2023-05-31"

type ClaudeRequest struct {
	*claude.ClaudeRequest
	AnthropicVersion string `json:"anthropic_version"`
}

func init() {
	CategoryMap["anthropic"] = Category{
		ChatComplete:              ConvertClaudeFromChatOpenai,
		ResponseChatComplete:      ConvertClaudeToChatOpenai,
		ResponseChatCompleteStrem: ClaudeChatCompleteStrem,
	}
}

func ConvertClaudeFromChatOpenai(request *types.ChatCompletionRequest) (any, *types.OpenAIErrorWithStatusCode) {
	rawRequest, err := claude.ConvertFromChatOpenai(request)
	if err != nil {
		return nil, err
	}

	claudeRequest := &ClaudeRequest{}
	claudeRequest.ClaudeRequest = rawRequest
	claudeRequest.AnthropicVersion = anthropicVersion

	// 删除model字段
	claudeRequest.Model = ""
	claudeRequest.Stream = false

	return claudeRequest, nil
}

func ConvertClaudeToChatOpenai(provider base.ProviderInterface, response *http.Response, request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	claudeResponse := &claude.ClaudeResponse{}
	err := json.NewDecoder(response.Body).Decode(claudeResponse)
	if err != nil {
		return nil, common.ErrorWrapper(err, "decode_response_failed", http.StatusInternalServerError)
	}

	return claude.ConvertToChatOpenai(provider, claudeResponse, request)
}

func ClaudeChatCompleteStrem(provider base.ProviderInterface, request *types.ChatCompletionRequest) requester.HandlerPrefix[string] {
	chatHandler := &claude.ClaudeStreamHandler{
		Usage:   provider.GetUsage(),
		Request: request,
		Prefix:  `{"type"`,
	}

	return chatHandler.HandlerStream
}
