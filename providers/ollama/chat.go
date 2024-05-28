package ollama

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/image"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/types"
	"strings"
)

type ollamaStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *OllamaProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &ChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(response, request)
}

func (p *OllamaProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &ollamaStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream(p.Requester, resp, chatHandler.handlerStream)
}

func (p *OllamaProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)

	// 获取请求头
	headers := p.GetRequestHeaders()

	ollamaRequest, errWithCode := convertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(ollamaRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *OllamaProvider) convertToChatOpenai(response *ChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	err := errorHandle(&response.OllamaError)
	if err != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *err,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	choices := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    response.Message.Role,
			Content: response.Message.Content,
		},
		FinishReason: types.FinishReasonStop,
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Model:   request.Model,
		Choices: []types.ChatCompletionChoice{choices},
		Usage: &types.Usage{
			PromptTokens:     response.PromptEvalCount,
			CompletionTokens: response.EvalCount,
			TotalTokens:      response.PromptEvalCount + response.EvalCount,
		},
	}

	*p.Usage = *openaiResponse.Usage

	return openaiResponse, nil
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) (*ChatRequest, *types.OpenAIErrorWithStatusCode) {
	ollamaRequest := &ChatRequest{
		Model:    request.Model,
		Stream:   request.Stream,
		Messages: make([]Message, 0, len(request.Messages)),
		Options: Option{
			Temperature: request.Temperature,
			TopP:        request.TopP,
			Seed:        request.Seed,
		},
	}

	for _, message := range request.Messages {
		ollamaMessage := Message{
			Role:    message.Role,
			Content: "",
		}

		openaiMessagePart := message.ParseContent()
		for _, openaiPart := range openaiMessagePart {
			if openaiPart.Type == types.ContentTypeText {
				ollamaMessage.Content += openaiPart.Text
			} else if openaiPart.Type == types.ContentTypeImageURL {
				_, data, err := image.GetImageFromUrl(openaiPart.ImageURL.URL)
				if err != nil {
					return nil, common.ErrorWrapper(err, "image_url_invalid", http.StatusBadRequest)
				}
				ollamaMessage.Images = append(ollamaMessage.Images, data)
			}
		}
		ollamaRequest.Messages = append(ollamaRequest.Messages, ollamaMessage)
	}

	return ollamaRequest, nil
}

// 转换为OpenAI聊天流式请求体
func (h *ollamaStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	if !strings.HasPrefix(string(*rawLine), "{") {
		*rawLine = nil
		return
	}

	var chatResponse ChatResponse
	err := json.Unmarshal(*rawLine, &chatResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	errWithCode := errorHandle(&chatResponse.OllamaError)
	if errWithCode != nil {
		errChan <- errWithCode
		return
	}

	choice := types.ChatCompletionStreamChoice{
		Index: 0,
	}

	if chatResponse.Message.Content != "" {
		choice.Delta = types.ChatCompletionStreamChoiceDelta{
			Role:    types.ChatMessageRoleAssistant,
			Content: chatResponse.Message.Content,
		}
	}

	if chatResponse.Done {
		choice.FinishReason = types.FinishReasonStop
	}

	if chatResponse.EvalCount > 0 {
		h.Usage.PromptTokens = chatResponse.PromptEvalCount
		h.Usage.CompletionTokens = chatResponse.EvalCount
		h.Usage.TotalTokens = h.Usage.PromptTokens + chatResponse.EvalCount
	}

	chatCompletion := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", utils.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: utils.GetTimestamp(),
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}

	responseBody, _ := json.Marshal(chatCompletion)
	dataChan <- string(responseBody)
}
