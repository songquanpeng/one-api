package mistral

import (
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type mistralStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *MistralProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &types.ChatCompletionResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	*p.Usage = *response.Usage

	return response, nil
}

func (p *MistralProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &mistralStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *MistralProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)

	// 获取请求头
	headers := p.GetRequestHeaders()

	mistralRequest := convertFromChatOpenai(request)

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(mistralRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *MistralChatCompletionRequest {
	request.ClearEmptyMessages()
	mistralRequest := &MistralChatCompletionRequest{
		Model:       request.Model,
		Messages:    make([]types.ChatCompletionMessage, 0, len(request.Messages)),
		Temperature: request.Temperature,
		MaxTokens:   request.MaxTokens,
		TopP:        request.TopP,
		N:           request.N,
		Stream:      request.Stream,
		Seed:        request.Seed,
	}

	for _, message := range request.Messages {
		mistralRequest.Messages = append(mistralRequest.Messages, types.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.StringContent(),
		})
	}

	if request.Tools != nil {
		mistralRequest.Tools = request.Tools
		mistralRequest.ToolChoice = "auto"
	}

	return mistralRequest
}

// 转换为OpenAI聊天流式请求体
func (h *mistralStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	*rawLine = (*rawLine)[6:]

	if string(*rawLine) == "[DONE]" {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	mistralResponse := &ChatCompletionStreamResponse{}
	err := json.Unmarshal(*rawLine, mistralResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	if mistralResponse.Usage != nil {
		*h.Usage = *mistralResponse.Usage
	} else {
		h.Usage.CompletionTokens += common.CountTokenText(mistralResponse.GetResponseText(), h.Request.Model)
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens
	}

	stop := false
	for _, choice := range mistralResponse.ChatCompletionStreamResponse.Choices {
		if choice.Delta.ToolCalls != nil {
			choices := choice.ConvertOpenaiStream()
			for _, newChoice := range choices {
				chatCompletionCopy := mistralResponse
				chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{newChoice}
				responseBody, _ := json.Marshal(chatCompletionCopy.ChatCompletionStreamResponse)
				dataChan <- string(responseBody)
			}
			stop = true
		}
	}

	if stop {
		return
	}

	responseBody, _ := json.Marshal(mistralResponse.ChatCompletionStreamResponse)
	dataChan <- string(responseBody)

}
