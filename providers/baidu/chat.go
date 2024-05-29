package baidu

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

type baiduStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *BaiduProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getBaiduChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	baiduResponse := &BaiduChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, baiduResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(baiduResponse, request)
}

func (p *BaiduProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getBaiduChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := &baiduStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *BaiduProvider) getBaiduChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_baidu_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()
	if request.Stream {
		headers["Accept"] = "text/event-stream"
	}

	baiduRequest := convertFromChatOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(baiduRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *BaiduProvider) convertToChatOpenai(response *BaiduChatResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.BaiduError)
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
			Role: "assistant",
		},
		FinishReason: types.FinishReasonStop,
	}

	if response.FunctionCall != nil {
		if request.Tools != nil {
			choice.Message.ToolCalls = []*types.ChatCompletionToolCalls{
				{
					Id:       response.Id,
					Type:     "function",
					Function: response.FunctionCall,
				},
			}
			choice.FinishReason = types.FinishReasonToolCalls
		} else {
			choice.Message.FunctionCall = response.FunctionCall
			choice.FinishReason = types.FinishReasonFunctionCall
		}
	} else {
		choice.Message.Content = response.Result
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.Id,
		Object:  "chat.completion",
		Model:   request.Model,
		Created: response.Created,
		Choices: []types.ChatCompletionChoice{choice},
		Usage:   response.Usage,
	}

	*p.Usage = *openaiResponse.Usage

	return
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *BaiduChatRequest {
	request.ClearEmptyMessages()
	baiduChatRequest := &BaiduChatRequest{
		Messages:        make([]BaiduMessage, 0, len(request.Messages)),
		Temperature:     request.Temperature,
		Stream:          request.Stream,
		TopP:            request.TopP,
		PenaltyScore:    request.FrequencyPenalty,
		Stop:            request.Stop,
		MaxOutputTokens: request.MaxTokens,
	}

	if request.ResponseFormat != nil {
		baiduChatRequest.ResponseFormat = request.ResponseFormat.Type

	}

	for _, message := range request.Messages {
		if message.Role == types.ChatMessageRoleSystem {
			baiduChatRequest.System = message.StringContent()
			continue
		} else if message.ToolCalls != nil {
			baiduChatRequest.Messages = append(baiduChatRequest.Messages, BaiduMessage{
				Role: types.ChatMessageRoleAssistant,
				FunctionCall: &types.ChatCompletionToolCallsFunction{
					Name:      *message.Name,
					Arguments: "{}",
				},
			})
		} else if message.Role == types.ChatMessageRoleFunction || message.Role == types.ChatMessageRoleTool {
			baiduChatRequest.Messages = append(baiduChatRequest.Messages, BaiduMessage{
				Role:    types.ChatMessageRoleUser,
				Content: "这是函数调用返回的内容，请回答之前的问题：\n" + message.StringContent(),
			})
		} else {
			baiduChatRequest.Messages = append(baiduChatRequest.Messages, BaiduMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}

	if request.Tools != nil {
		functions := make([]*types.ChatCompletionFunction, 0, len(request.Tools))
		for _, tool := range request.Tools {
			functions = append(functions, &tool.Function)
		}
		baiduChatRequest.Functions = functions
	} else if request.Functions != nil {
		baiduChatRequest.Functions = request.Functions
	}

	return baiduChatRequest
}

// 转换为OpenAI聊天流式请求体
func (h *baiduStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	var baiduResponse BaiduChatStreamResponse
	err := json.Unmarshal(*rawLine, &baiduResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := errorHandle(&baiduResponse.BaiduError)
	if aiError != nil {
		errChan <- aiError
		return
	}

	h.convertToOpenaiStream(&baiduResponse, dataChan)

	if baiduResponse.IsEnd {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}
}

func (h *baiduStreamHandler) convertToOpenaiStream(baiduResponse *BaiduChatStreamResponse, dataChan chan string) {
	choice := types.ChatCompletionStreamChoice{
		Index: 0,
		Delta: types.ChatCompletionStreamChoiceDelta{
			Role: "assistant",
		},
	}

	if baiduResponse.FunctionCall != nil {
		if h.Request.Tools != nil {
			choice.Delta.ToolCalls = []*types.ChatCompletionToolCalls{
				{
					Id:       baiduResponse.Id,
					Type:     "function",
					Function: baiduResponse.FunctionCall,
				},
			}
			choice.FinishReason = types.FinishReasonToolCalls
		} else {
			choice.Delta.FunctionCall = baiduResponse.FunctionCall
			choice.FinishReason = types.FinishReasonFunctionCall
		}
	} else {
		choice.Delta.Content = baiduResponse.Result
		if baiduResponse.IsEnd {
			choice.FinishReason = types.FinishReasonStop
		}
	}

	chatCompletion := types.ChatCompletionStreamResponse{
		ID:      baiduResponse.Id,
		Object:  "chat.completion.chunk",
		Created: baiduResponse.Created,
		Model:   h.Request.Model,
	}

	if baiduResponse.FunctionCall == nil {
		chatCompletion.Choices = []types.ChatCompletionStreamChoice{choice}
		responseBody, _ := json.Marshal(chatCompletion)
		dataChan <- string(responseBody)
	} else {
		choices := choice.ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := chatCompletion
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	}

	h.Usage.TotalTokens = baiduResponse.Usage.TotalTokens
	h.Usage.PromptTokens = baiduResponse.Usage.PromptTokens
	h.Usage.CompletionTokens += baiduResponse.Usage.CompletionTokens
}
