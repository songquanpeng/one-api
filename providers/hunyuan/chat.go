package hunyuan

import (
	"encoding/json"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type tunyuanStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *HunyuanProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	tunyuanChatResponse := &ChatCompletionsResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, tunyuanChatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(tunyuanChatResponse, request)
}

func (p *HunyuanProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &tunyuanStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *HunyuanProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	action, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	tunyuanRequest := convertFromChatOpenai(request)
	req, errWithCode := p.sign(tunyuanRequest, action, http.MethodPost)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return req, nil
}

func (p *HunyuanProvider) convertToChatOpenai(response *ChatCompletionsResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.Response.HunyuanResponseError)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	txResponse := response.Response

	openaiResponse = &types.ChatCompletionResponse{
		Object:  "chat.completion",
		Created: txResponse.Created,
		Usage: &types.Usage{
			PromptTokens:     txResponse.Usage.PromptTokens,
			CompletionTokens: txResponse.Usage.CompletionTokens,
			TotalTokens:      txResponse.Usage.TotalTokens,
		},
		Model: request.Model,
	}

	for _, choice := range txResponse.Choices {
		openaiResponse.Choices = append(openaiResponse.Choices, types.ChatCompletionChoice{
			Index:        0,
			Message:      types.ChatCompletionMessage{Role: choice.Message.Role, Content: choice.Message.Content},
			FinishReason: choice.FinishReason,
		})

	}

	*p.Usage = *openaiResponse.Usage

	return
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *ChatCompletionsRequest {
	request.ClearEmptyMessages()

	messages := make([]*Message, 0, len(request.Messages))
	for _, message := range request.Messages {
		messages = append(messages, &Message{
			Content: message.StringContent(),
			Role:    message.Role,
		})
	}

	return &ChatCompletionsRequest{
		Model:       request.Model,
		Messages:    messages,
		Stream:      request.Stream,
		TopP:        &request.TopP,
		Temperature: &request.Temperature,
	}
}

// 转换为OpenAI聊天流式请求体
func (h *tunyuanStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data:") {
		*rawLine = nil
		return
	}

	// 去除前缀
	*rawLine = (*rawLine)[5:]

	var tunyuanChatResponse ChatCompletionsResponseParams
	err := json.Unmarshal(*rawLine, &tunyuanChatResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&tunyuanChatResponse.HunyuanResponseError)
	if aiError != nil {
		errChan <- aiError
		return
	}

	h.convertToOpenaiStream(&tunyuanChatResponse, dataChan)

}

func (h *tunyuanStreamHandler) convertToOpenaiStream(tunyuanChatResponse *ChatCompletionsResponseParams, dataChan chan string) {
	streamResponse := types.ChatCompletionStreamResponse{
		Object:  "chat.completion.chunk",
		Created: tunyuanChatResponse.Created,
		Model:   h.Request.Model,
	}

	for _, choice := range tunyuanChatResponse.Choices {
		streamResponse.Choices = append(streamResponse.Choices, types.ChatCompletionStreamChoice{
			FinishReason: choice.FinishReason,
			Delta: types.ChatCompletionStreamChoiceDelta{
				Role:    choice.Delta.Role,
				Content: choice.Delta.Content,
			},
			Index: 0,
		})
	}

	responseBody, _ := json.Marshal(streamResponse)
	dataChan <- string(responseBody)

	*h.Usage = types.Usage{
		PromptTokens:     tunyuanChatResponse.Usage.PromptTokens,
		CompletionTokens: tunyuanChatResponse.Usage.CompletionTokens,
		TotalTokens:      tunyuanChatResponse.Usage.TotalTokens,
	}
}
