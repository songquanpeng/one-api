package cloudflareAI

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/types"
	"strings"
)

type CloudflareAIStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *CloudflareAIProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	chatResponse := &ChatRespone{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, chatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(chatResponse, request)
}

func (p *CloudflareAIProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &CloudflareAIStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *CloudflareAIProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(request.Model)
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

func (p *CloudflareAIProvider) convertToChatOpenai(response *ChatRespone, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	err := errorHandle(&response.CloudflareAIError)
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
				Content: response.Result.Response,
			},
			FinishReason: types.FinishReasonStop,
		}},
	}

	completionTokens := common.CountTokenText(response.Result.Response, request.Model)

	p.Usage.CompletionTokens = completionTokens
	p.Usage.TotalTokens = p.Usage.PromptTokens + completionTokens
	openaiResponse.Usage = p.Usage

	return
}

func (p *CloudflareAIProvider) convertFromChatOpenai(request *types.ChatCompletionRequest) *ChatRequest {
	request.ClearEmptyMessages()
	chatRequest := &ChatRequest{
		Stream:    request.Stream,
		MaxTokens: request.MaxTokens,
		Messages:  make([]Message, 0, len(request.Messages)),
	}

	for _, message := range request.Messages {
		chatRequest.Messages = append(chatRequest.Messages, Message{
			Role:    message.Role,
			Content: message.StringContent(),
		})
	}

	return chatRequest
}

// 转换为OpenAI聊天流式请求体
func (h *CloudflareAIStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data: 或者 meta:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	*rawLine = (*rawLine)[6:]

	if strings.HasPrefix(string(*rawLine), "[DONE]") {
		h.convertToOpenaiStream(nil, dataChan, true)
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	chatResponse := &ChatResult{}
	err := json.Unmarshal(*rawLine, chatResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	h.convertToOpenaiStream(chatResponse, dataChan, false)
}

func (h *CloudflareAIStreamHandler) convertToOpenaiStream(chatResponse *ChatResult, dataChan chan string, isStop bool) {
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

	if isStop {
		choice.FinishReason = types.FinishReasonStop
	} else {
		choice.Delta.Content = chatResponse.Response

		h.Usage.CompletionTokens += common.CountTokenText(chatResponse.Response, h.Request.Model)
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens
	}

	streamResponse.Choices = []types.ChatCompletionStreamChoice{choice}
	responseBody, _ := json.Marshal(streamResponse)
	dataChan <- string(responseBody)

}
