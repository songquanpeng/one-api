package baichuan

import (
	"net/http"
	"one-api/common"
	"one-api/providers/openai"
	"one-api/types"
	"strings"
)

func (baichuanResponse *BaichuanChatResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if baichuanResponse.Error.Message != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: baichuanResponse.Error,
			StatusCode:  resp.StatusCode,
		}

		return
	}

	OpenAIResponse = types.ChatCompletionResponse{
		ID:      baichuanResponse.ID,
		Object:  baichuanResponse.Object,
		Created: baichuanResponse.Created,
		Model:   baichuanResponse.Model,
		Choices: baichuanResponse.Choices,
		Usage:   baichuanResponse.Usage,
	}

	return
}

// 获取聊天请求体
func (p *BaichuanProvider) getChatRequestBody(request *types.ChatCompletionRequest) *BaichuanChatRequest {
	messages := make([]BaichuanMessage, 0, len(request.Messages))
	for i := 0; i < len(request.Messages); i++ {
		message := request.Messages[i]
		if message.Role == "system" || message.Role == "assistant" {
			message.Role = "assistant"
		} else {
			message.Role = "user"
		}
		messages = append(messages, BaichuanMessage{
			Content: message.StringContent(),
			Role:    strings.ToLower(message.Role),
		})
	}

	return &BaichuanChatRequest{
		Model:       request.Model,
		Messages:    messages,
		Stream:      request.Stream,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		TopK:        request.N,
	}
}

// 聊天
func (p *BaichuanProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody := p.getChatRequestBody(request)

	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		openAIProviderChatStreamResponse := &openai.OpenAIProviderChatStreamResponse{}
		var textResponse string
		errWithCode, textResponse = p.SendStreamRequest(req, openAIProviderChatStreamResponse)
		if errWithCode != nil {
			return
		}

		usage = &types.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: common.CountTokenText(textResponse, request.Model),
			TotalTokens:      promptTokens + common.CountTokenText(textResponse, request.Model),
		}

	} else {
		baichuanResponse := &BaichuanChatResponse{}
		errWithCode = p.SendRequest(req, baichuanResponse, false)
		if errWithCode != nil {
			return
		}

		usage = baichuanResponse.Usage
	}
	return
}
