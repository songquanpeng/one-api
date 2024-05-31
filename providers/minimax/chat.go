package minimax

import (
	"encoding/json"
	"errors"
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type minimaxStreamHandler struct {
	Usage       *types.Usage
	Request     *types.ChatCompletionRequest
	LastContent string
}

func (p *MiniMaxProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &MiniMaxChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(response, request)
}

func (p *MiniMaxProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &minimaxStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *MiniMaxProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(errors.New("API KEY is filled in incorrectly"), "invalid_minimax_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	zhipuRequest := convertFromChatOpenai(request)

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(zhipuRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *MiniMaxProvider) convertToChatOpenai(response *MiniMaxChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.MiniMaxBaseResp.BaseResp)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.ID,
		Object:  "chat.completion",
		Created: response.Created,
		Model:   response.Model,
		Choices: make([]types.ChatCompletionChoice, 0, len(response.Choices)),
	}

	for _, choice := range response.Choices {
		openaiChoice := types.ChatCompletionChoice{
			FinishReason: convertFinishReason(choice.FinishReason),
		}
		if choice.Messages[0].FunctionCall != nil {
			if request.Functions != nil {
				openaiChoice.Message.FunctionCall = choice.Messages[0].FunctionCall
			} else {
				openaiChoice.Message.ToolCalls = append(openaiChoice.Message.ToolCalls, &types.ChatCompletionToolCalls{
					Type:     types.ChatMessageRoleFunction,
					Function: choice.Messages[0].FunctionCall,
				})
			}
		} else {
			openaiChoice.Message.Role = choice.Messages[0].SenderName
			openaiChoice.Message.Content = choice.Messages[0].Text
		}
		openaiResponse.Choices = append(openaiResponse.Choices, openaiChoice)
	}

	if response.Usage.TotalTokens < p.Usage.PromptTokens {
		p.Usage.PromptTokens = response.Usage.TotalTokens
	}
	p.Usage.TotalTokens = response.Usage.TotalTokens
	p.Usage.CompletionTokens = response.Usage.TotalTokens - p.Usage.PromptTokens

	openaiResponse.Usage = p.Usage

	return
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *MiniMaxChatRequest {
	var botSettings []MiniMaxBotSetting
	var messges []MiniMaxChatMessage
	request.ClearEmptyMessages()
	for _, message := range request.Messages {
		if message.Role == types.ChatMessageRoleSystem {
			botSettings = append(botSettings, MiniMaxBotSetting{
				BotName: types.ChatMessageRoleAssistant,
				Content: message.StringContent(),
			})
			continue
		}
		miniMessage := MiniMaxChatMessage{
			Text: message.StringContent(),
		}

		// 如果role为function， 则需要在前面一条记录添加function_call，如果没有消息，则添加一个message
		if message.ToolCalls != nil {
			miniMessage.FunctionCall = &types.ChatCompletionToolCallsFunction{
				Name:      message.ToolCalls[0].Function.Name,
				Arguments: message.ToolCalls[0].Function.Arguments,
			}
		} else if message.Role == types.ChatMessageRoleFunction {
			if len(messges) == 0 {
				messges = append(messges, MiniMaxChatMessage{
					SenderType: "USER",
					SenderName: types.ChatMessageRoleUser,
				})
			}

			messges[len(messges)-1].FunctionCall = &types.ChatCompletionToolCallsFunction{
				Name:      "function",
				Arguments: "arguments",
			}
		}

		miniMessage.SenderType, miniMessage.SenderName = convertRole(message.Role)

		messges = append(messges, miniMessage)
	}

	if len(botSettings) == 0 {
		botSettings = append(botSettings, defaultBot())
	}

	miniRequest := &MiniMaxChatRequest{
		Model:            request.Model,
		Messages:         messges,
		Stream:           request.Stream,
		Temperature:      request.Temperature,
		TopP:             request.TopP,
		TokensToGenerate: request.MaxTokens,
		BotSetting:       botSettings,
		ReplyConstraints: defaultReplyConstraints(),
	}

	if request.Functions != nil {
		miniRequest.Functions = request.Functions
	} else if request.Tools != nil {
		miniRequest.Functions = make([]*types.ChatCompletionFunction, 0, len(request.Tools))
		for _, tool := range request.Tools {
			miniRequest.Functions = append(miniRequest.Functions, &tool.Function)
		}
	}
	return miniRequest
}

// 转换为OpenAI聊天流式请求体
func (h *minimaxStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data: 或者 meta:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	*rawLine = (*rawLine)[6:]

	miniResponse := &MiniMaxChatResponse{}
	err := json.Unmarshal(*rawLine, miniResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&miniResponse.BaseResp)
	if aiError != nil {
		errChan <- aiError
		return
	}

	choice := miniResponse.Choices[0]

	if choice.Messages[0].FunctionCall != nil && choice.FinishReason == "" {
		*rawLine = nil
		return
	}

	h.convertToOpenaiStream(miniResponse, dataChan)
}

func (h *minimaxStreamHandler) convertToOpenaiStream(miniResponse *MiniMaxChatResponse, dataChan chan string) {
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      miniResponse.RequestID,
		Object:  "chat.completion.chunk",
		Created: miniResponse.Created,
		Model:   h.Request.Model,
	}

	miniChoice := miniResponse.Choices[0]
	openaiChoice := types.ChatCompletionStreamChoice{}

	if miniChoice.Messages[0].FunctionCall == nil && miniChoice.FinishReason != "" {
		streamResponse.ID = miniResponse.ID
		openaiChoice.FinishReason = convertFinishReason(miniChoice.FinishReason)
		dataChan <- h.getResponseString(&streamResponse, &openaiChoice)
		if miniResponse.Usage != nil {
			h.handleUsage(miniResponse)
		}
		return
	}

	openaiChoice.Delta = types.ChatCompletionStreamChoiceDelta{
		Role: miniChoice.Messages[0].SenderName,
	}

	if miniChoice.Messages[0].FunctionCall != nil {
		h.handleFunctionCall(&miniChoice, &openaiChoice)
		convertChoices := openaiChoice.ConvertOpenaiStream()
		for _, convertChoice := range convertChoices {
			chatCompletionCopy := streamResponse
			dataChan <- h.getResponseString(&chatCompletionCopy, &convertChoice)
		}

	} else {
		openaiChoice.Delta.Content = miniChoice.Messages[0].Text
		dataChan <- h.getResponseString(&streamResponse, &openaiChoice)
	}

	if miniResponse.Usage != nil {
		h.handleUsage(miniResponse)
	} else {
		h.Usage.CompletionTokens += common.CountTokenText(miniChoice.Messages[0].Text, h.Request.Model)
		h.Usage.TotalTokens = h.Usage.PromptTokens + h.Usage.CompletionTokens
	}
}

func (h *minimaxStreamHandler) handleFunctionCall(choice *Choice, openaiChoice *types.ChatCompletionStreamChoice) {
	if h.Request.Functions != nil {
		openaiChoice.Delta.FunctionCall = choice.Messages[0].FunctionCall
	} else {
		openaiChoice.Delta.ToolCalls = append(openaiChoice.Delta.ToolCalls, &types.ChatCompletionToolCalls{
			Type:     types.ChatMessageRoleFunction,
			Function: choice.Messages[0].FunctionCall,
		})
	}
}

func (h *minimaxStreamHandler) getResponseString(streamResponse *types.ChatCompletionStreamResponse, openaiChoice *types.ChatCompletionStreamChoice) string {
	streamResponse.Choices = []types.ChatCompletionStreamChoice{*openaiChoice}
	responseBody, _ := json.Marshal(streamResponse)
	return string(responseBody)
}

func (h *minimaxStreamHandler) handleUsage(miniResponse *MiniMaxChatResponse) {
	if miniResponse.Usage.TotalTokens < h.Usage.PromptTokens {
		h.Usage.PromptTokens = miniResponse.Usage.TotalTokens
	}
	h.Usage.TotalTokens = miniResponse.Usage.TotalTokens
	h.Usage.CompletionTokens = miniResponse.Usage.TotalTokens - h.Usage.PromptTokens
}
