package cohere

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/common/utils"
	"one-api/providers/base"
	"one-api/types"
	"strings"
)

type CohereStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *CohereProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	cohereResponse := &CohereResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, cohereResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return ConvertToChatOpenai(p, cohereResponse, request)
}

func (p *CohereProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &CohereStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream(p.Requester, resp, chatHandler.HandlerStream)
}

func (p *CohereProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_cohere_config", http.StatusInternalServerError)
	}

	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	cohereRequest, errWithCode := ConvertFromChatOpenai(request)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(cohereRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func ConvertFromChatOpenai(request *types.ChatCompletionRequest) (*CohereRequest, *types.OpenAIErrorWithStatusCode) {
	request.ClearEmptyMessages()
	cohereRequest := CohereRequest{
		Model:            request.Model,
		MaxTokens:        request.MaxTokens,
		Temperature:      request.Temperature,
		Stream:           request.Stream,
		P:                request.TopP,
		K:                request.N,
		Seed:             request.Seed,
		StopSequences:    request.Stop,
		FrequencyPenalty: request.FrequencyPenalty,
		PresencePenalty:  request.PresencePenalty,
	}

	msgLen := len(request.Messages) - 1

	for index, message := range request.Messages {
		if index == msgLen {
			cohereRequest.Message = message.StringContent()
		} else {
			cohereRequest.ChatHistory = append(cohereRequest.ChatHistory, ChatHistory{
				Role:    convertRole(message.Role),
				Message: message.StringContent(),
			})
		}

	}

	return &cohereRequest, nil
}

func ConvertToChatOpenai(provider base.ProviderInterface, response *CohereResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.CohereError)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	choice := types.ChatCompletionChoice{
		Index: 0,
		Message: types.ChatCompletionMessage{
			Role:    types.ChatMessageRoleAssistant,
			Content: response.Text,
		},
		FinishReason: types.FinishReasonStop,
	}
	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.GenerationID,
		Object:  "chat.completion",
		Created: utils.GetTimestamp(),
		Choices: []types.ChatCompletionChoice{choice},
		Model:   request.Model,
		Usage:   &types.Usage{},
	}
	*openaiResponse.Usage = usageHandle(&response.Meta.BilledUnits)

	usage := provider.GetUsage()
	*usage = *openaiResponse.Usage

	return openaiResponse, nil
}

// 转换为OpenAI聊天流式请求体
func (h *CohereStreamHandler) HandlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "{") {
		*rawLine = nil
		return
	}

	var cohereResponse CohereStreamResponse
	err := json.Unmarshal(*rawLine, &cohereResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	if cohereResponse.EventType != "text-generation" && cohereResponse.EventType != "stream-end" {
		*rawLine = nil
		return
	}

	h.convertToOpenaiStream(&cohereResponse, dataChan)
}

func (h *CohereStreamHandler) convertToOpenaiStream(cohereResponse *CohereStreamResponse, dataChan chan string) {
	choice := types.ChatCompletionStreamChoice{
		Index: 0,
	}

	if cohereResponse.EventType == "stream-end" {
		choice.FinishReason = types.FinishReasonStop
		*h.Usage = usageHandle(&cohereResponse.Response.Meta.BilledUnits)
	} else {
		choice.Delta = types.ChatCompletionStreamChoiceDelta{
			Role:    types.ChatMessageRoleAssistant,
			Content: cohereResponse.Text,
		}

		h.Usage.CompletionTokens += common.CountTokenText(cohereResponse.Text, h.Request.Model)
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens
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

func usageHandle(token *Tokens) types.Usage {
	usage := types.Usage{
		PromptTokens:     token.InputTokens,
		CompletionTokens: token.OutputTokens + token.SearchUnits + token.Classifications,
	}
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	return usage
}
