package palm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/types"
	"strings"
)

type palmStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *PalmProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	palmResponse := &PaLMChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, palmResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(palmResponse, request)
}

func (p *PalmProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &palmStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *PalmProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_palm_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	palmRequest := convertFromChatOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(palmRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *PalmProvider) convertToChatOpenai(response *PaLMChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.PaLMErrorResponse)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		Choices: make([]types.ChatCompletionChoice, 0, len(response.Candidates)),
		Model:   request.Model,
	}
	for i, candidate := range response.Candidates {
		choice := types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatCompletionMessage{
				Role:    "assistant",
				Content: candidate.Content,
			},
			FinishReason: "stop",
		}
		openaiResponse.Choices = append(openaiResponse.Choices, choice)
	}

	completionTokens := common.CountTokenText(response.Candidates[0].Content, request.Model)
	response.Usage.CompletionTokens = completionTokens
	response.Usage.TotalTokens = response.Usage.PromptTokens + completionTokens

	openaiResponse.Usage = response.Usage

	*p.Usage = *response.Usage

	return
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *PaLMChatRequest {
	palmRequest := PaLMChatRequest{
		Prompt: PaLMPrompt{
			Messages: make([]PaLMChatMessage, 0, len(request.Messages)),
		},
		Temperature:    request.Temperature,
		CandidateCount: request.N,
		TopP:           request.TopP,
		TopK:           request.MaxTokens,
	}
	for _, message := range request.Messages {
		palmMessage := PaLMChatMessage{
			Content: message.StringContent(),
		}
		if message.Role == "user" {
			palmMessage.Author = "0"
		} else {
			palmMessage.Author = "1"
		}
		palmRequest.Prompt.Messages = append(palmRequest.Prompt.Messages, palmMessage)
	}
	return &palmRequest
}

// 转换为OpenAI聊天流式请求体
func (h *palmStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	var palmChatResponse PaLMChatResponse
	err := json.Unmarshal(*rawLine, &palmChatResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&palmChatResponse.PaLMErrorResponse)
	if aiError != nil {
		errChan <- aiError
		return
	}

	h.convertToOpenaiStream(&palmChatResponse, dataChan)

}

func (h *palmStreamHandler) convertToOpenaiStream(palmChatResponse *PaLMChatResponse, dataChan chan string) {
	var choice types.ChatCompletionStreamChoice
	if len(palmChatResponse.Candidates) > 0 {
		choice.Delta.Content = palmChatResponse.Candidates[0].Content
	}
	choice.FinishReason = types.FinishReasonStop

	streamResponse := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
		Object:  "chat.completion.chunk",
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
		Created: utils.GetTimestamp(),
	}

	responseBody, _ := json.Marshal(streamResponse)
	dataChan <- string(responseBody)

	h.Usage.CompletionTokens += common.CountTokenText(palmChatResponse.Candidates[0].Content, h.Request.Model)
	h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens
}
