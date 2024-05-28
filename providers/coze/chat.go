package coze

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/types"
	"strings"
)

type CozeStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *CozeProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	chatResponse := &CozeResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, chatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(chatResponse, request)
}

func (p *CozeProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &CozeStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *CozeProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_cloudflare_ai_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()
	chatRequest := p.convertFromChatOpenai(request)

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(chatRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *CozeProvider) convertToChatOpenai(response *CozeResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	err := errorHandle(&response.CozeStatus)
	if err != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *err,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Model:   request.Model,
		Choices: []types.ChatCompletionChoice{{
			Index: 0,
			Message: types.ChatCompletionMessage{
				Role:    types.ChatMessageRoleAssistant,
				Content: response.String(),
			},
			FinishReason: types.FinishReasonStop,
		}},
	}

	p.Usage.CompletionTokens = common.CountTokenText(response.String(), request.Model)
	p.Usage.TotalTokens = p.Usage.CompletionTokens + p.Usage.PromptTokens
	openaiResponse.Usage = p.Usage

	return
}

func (p *CozeProvider) convertFromChatOpenai(request *types.ChatCompletionRequest) *CozeRequest {
	model := strings.TrimPrefix(request.Model, "coze-")
	chatRequest := &CozeRequest{
		Stream: request.Stream,
		BotID:  model,
		User:   "OneAPI",
	}
	msgLen := len(request.Messages) - 1

	for index, message := range request.Messages {
		if index == msgLen {
			chatRequest.Query = message.StringContent()
		} else {
			chatRequest.ChatHistory = append(chatRequest.ChatHistory, CozeMessage{
				Role:        convertRole(message.Role),
				Content:     message.StringContent(),
				ContentType: "text",
			})
		}

	}

	return chatRequest
}

// 转换为OpenAI聊天流式请求体
func (h *CozeStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data: 或者 meta:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data:") {
		*rawLine = nil
		return
	}

	*rawLine = (*rawLine)[5:]

	chatResponse := &CozeStreamResponse{}
	err := json.Unmarshal(*rawLine, chatResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	if chatResponse.Event == "done" {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	if chatResponse.Event != "message" || chatResponse.Message.Type != "answer" {
		*rawLine = nil
		return
	}

	h.convertToOpenaiStream(chatResponse, dataChan)
}

func (h *CozeStreamHandler) convertToOpenaiStream(chatResponse *CozeStreamResponse, dataChan chan string) {
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   h.Request.Model,
	}

	choice := types.ChatCompletionStreamChoice{
		Index: 0,
		Delta: types.ChatCompletionStreamChoiceDelta{
			Role:    types.ChatMessageRoleAssistant,
			Content: "",
		},
	}

	if chatResponse.IsFinish {
		choice.FinishReason = types.FinishReasonStop
	} else {
		choice.Delta.Content = chatResponse.Message.Content
		h.Usage.CompletionTokens += common.CountTokenText(chatResponse.Message.Content, h.Request.Model)
		h.Usage.TotalTokens = h.Usage.CompletionTokens + h.Usage.PromptTokens
	}

	streamResponse.Choices = []types.ChatCompletionStreamChoice{choice}
	responseBody, _ := json.Marshal(streamResponse)
	dataChan <- string(responseBody)

}
