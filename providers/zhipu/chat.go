package zhipu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
	"strings"
)

type zhipuStreamHandler struct {
	Usage   *types.Usage
	Request *types.ChatCompletionRequest
}

func (p *ZhipuProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (*types.ChatCompletionResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getChatRequest(request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	zhipuChatResponse := &ZhipuResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, zhipuChatResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToChatOpenai(zhipuChatResponse, request)
}

func (p *ZhipuProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[types.ChatCompletionStreamResponse], *types.OpenAIErrorWithStatusCode) {
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

	chatHandler := &zhipuStreamHandler{
		Usage:   p.Usage,
		Request: request,
	}

	return requester.RequestStream[types.ChatCompletionStreamResponse](p.Requester, resp, chatHandler.handlerStream)
}

func (p *ZhipuProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(common.RelayModeChatCompletions)
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
		fullRequestURL += "/sse-invoke"
	} else {
		fullRequestURL += "/invoke"
	}

	zhipuRequest := convertFromChatOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(zhipuRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	return req, nil
}

func (p *ZhipuProvider) convertToChatOpenai(response *ZhipuResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	error := errorHandle(response)
	if error != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *error,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.Data.TaskId,
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Model:   request.Model,
		Choices: make([]types.ChatCompletionChoice, 0, len(response.Data.Choices)),
		Usage:   &response.Data.Usage,
	}
	for i, choice := range response.Data.Choices {
		openaiChoice := types.ChatCompletionChoice{
			Index: i,
			Message: types.ChatCompletionMessage{
				Role:    choice.Role,
				Content: strings.Trim(choice.Content, "\""),
			},
			FinishReason: "",
		}
		if i == len(response.Data.Choices)-1 {
			openaiChoice.FinishReason = "stop"
		}
		openaiResponse.Choices = append(openaiResponse.Choices, openaiChoice)
	}

	*p.Usage = response.Data.Usage

	return
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *ZhipuRequest {
	messages := make([]ZhipuMessage, 0, len(request.Messages))
	for _, message := range request.Messages {
		if message.Role == "system" {
			messages = append(messages, ZhipuMessage{
				Role:    "system",
				Content: message.StringContent(),
			})
			messages = append(messages, ZhipuMessage{
				Role:    "user",
				Content: "Okay",
			})
		} else {
			messages = append(messages, ZhipuMessage{
				Role:    message.Role,
				Content: message.StringContent(),
			})
		}
	}
	return &ZhipuRequest{
		Prompt:      messages,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Incremental: false,
	}
}

// 转换为OpenAI聊天流式请求体
func (h *zhipuStreamHandler) handlerStream(rawLine *[]byte, isFinished *bool, response *[]types.ChatCompletionStreamResponse) error {
	// 如果rawLine 前缀不为data: 或者 meta:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data:") && !strings.HasPrefix(string(*rawLine), "meta:") {
		*rawLine = nil
		return nil
	}

	if strings.HasPrefix(string(*rawLine), "meta:") {
		*rawLine = (*rawLine)[5:]
		var zhipuStreamMetaResponse ZhipuStreamMetaResponse
		err := json.Unmarshal(*rawLine, &zhipuStreamMetaResponse)
		if err != nil {
			return common.ErrorToOpenAIError(err)
		}
		*isFinished = true
		return h.handlerMeta(&zhipuStreamMetaResponse, response)
	}

	*rawLine = (*rawLine)[5:]
	return h.convertToOpenaiStream(string(*rawLine), response)
}

func (h *zhipuStreamHandler) convertToOpenaiStream(content string, response *[]types.ChatCompletionStreamResponse) error {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = content
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}

	*response = append(*response, streamResponse)

	return nil
}

func (h *zhipuStreamHandler) handlerMeta(zhipuResponse *ZhipuStreamMetaResponse, response *[]types.ChatCompletionStreamResponse) error {
	var choice types.ChatCompletionStreamChoice
	choice.Delta.Content = ""
	choice.FinishReason = types.FinishReasonStop
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      zhipuResponse.RequestId,
		Object:  "chat.completion.chunk",
		Created: common.GetTimestamp(),
		Model:   h.Request.Model,
		Choices: []types.ChatCompletionStreamChoice{choice},
	}

	*response = append(*response, streamResponse)

	*h.Usage = zhipuResponse.Usage

	return nil
}
