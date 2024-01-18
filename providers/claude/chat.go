package claude

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type claudeStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *ClaudeProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	claudeResponse := &ClaudeResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, claudeResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(claudeResponse, request)
}

func (p *ClaudeProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[types.ChatCompletionStreamResponse], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &claudeStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[types.ChatCompletionStreamResponse](p.Requester, resp, chatHandler.handlerStream)
}

func (p *ClaudeProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(common.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_claude_config", http.StatusInternalServerError)
	}

	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	claudeRequest := convertFromChatOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(claudeRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *ClaudeRequest {
	claudeRequest := ClaudeRequest{
		Model:             request.Model,
		Prompt:            "",
		MaxTokensToSample: request.MaxTokens,
		StopSequences:     nil,
		Temperature:       request.Temperature,
		TopP:              request.TopP,
		Stream:            request.Stream,
	}
	if claudeRequest.MaxTokensToSample == 0 {
		claudeRequest.MaxTokensToSample = 1000000
	}
	prompt := ""
	for _, message := range request.Messages {
		if message.Role == "user" {
			prompt += fmt.Sprintf("\n\nHuman: %s", message.Content)
		} else if message.Role == "assistant" {
			prompt += fmt.Sprintf("\n\nAssistant: %s", message.Content)
		} else if message.Role == "system" {
			if prompt == "" {
				prompt = message.StringContent()
			}
		}
	}
	prompt += "\n\nAssistant:"
	claudeRequest.Prompt = prompt
	return &claudeRequest
}

func (p *ClaudeProvider) convertToChatOpenai(response *ClaudeResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	error := errorHandle(&response.ClaudeResponseError)
	if error != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *error,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    "assistant",
			Content: strings.TrimPrefix(response.Completion, " "),
			Name:    nil,
		},
		FinishReason: stopReasonClaude2OpenAI(response.StopReason),
	}
	openaiResponse = &types.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
		Model:   response.Model,
	}

	completionTokens := common.CountTokenText(response.Completion, response.Model)
	response.Usage.CompletionTokens = completionTokens
	response.Usage.TotalTokens = response.Usage.PromptTokens + completionTokens

	openaiResponse.Usage = response.Usage

	*p.Usage = *response.Usage

	return openaiResponse, nil
}

// 转换为OpenAI聊天流式请求体
func (h *claudeStreamHandler) handlerStream(rawLine *[]byte, isFinished *bool, response *[]types.ChatCompletionStreamResponse) error {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), `data: {"type": "completion"`) {
		*rawLine = nil
		return nil
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	var claudeResponse *ClaudeResponse
	err := json.Unmarshal(*rawLine, claudeResponse)
	if err != nil {
		return common.ErrorToOpenAIError(err)
	}

	if claudeResponse.StopReason == "stop_sequence" {
		*isFinished = true
	}

	return h.convertToOpenaiStream(claudeResponse, response)
}

func (h *claudeStreamHandler) convertToOpenaiStream(claudeResponse *ClaudeResponse, response *[]types.ChatCompletionStreamResponse) error {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = claudeResponse.Completion
	finishReason := stopReasonClaude2OpenAI(claudeResponse.StopReason)
	if finishReason != "null" {
		choice.FinishReason = &finishReason
	}
	chatCompletion := types.ChatCompletionStreamResponse{
		Object:  "chat.completion.chunk",
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}

	*response = append(*response, chatCompletion)

	h.Usage.PromptTokens += common.CountTokenText(claudeResponse.Completion, h.Request.Model)

	return nil
}
