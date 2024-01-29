package zhipu

import (
	"encoding/json"
	"io"
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/types"
	"strings"
	"time"
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

func (p *ZhipuProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
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

	return requester.RequestStream[string](p.Requester, resp, chatHandler.handlerStream)
}

func (p *ZhipuProvider) getChatRequest(request *types.ChatCompletionRequest) (*http.Request, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(common.RelayModeChatCompletions)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url, request.Model)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_zhipu_config", http.StatusInternalServerError)
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

func (p *ZhipuProvider) convertToChatOpenai(response *ZhipuResponse, request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	error := errorHandle(&response.Error)
	if error != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *error,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.ChatCompletionResponse{
		ID:      response.ID,
		Object:  "chat.completion",
		Created: response.Created,
		Model:   response.Model,
		Choices: response.Choices,
		Usage:   response.Usage,
	}

	*p.Usage = *response.Usage

	return
}

func convertFromChatOpenai(request *types.ChatCompletionRequest) *ZhipuRequest {
	for i := range request.Messages {
		request.Messages[i].Role = convertRole(request.Messages[i].Role)
	}

	zhipuRequest := &ZhipuRequest{
		Model:       request.Model,
		Messages:    request.Messages,
		Stream:      request.Stream,
		Temperature: request.Temperature,
		TopP:        convertTopP(request.TopP),
		MaxTokens:   request.MaxTokens,
		Stop:        request.Stop,
		ToolChoice:  request.ToolChoice,
	}

	if request.Functions != nil {
		zhipuRequest.Tools = make([]ZhipuTool, 0, len(request.Functions))
		for _, function := range request.Functions {
			zhipuRequest.Tools = append(zhipuRequest.Tools, ZhipuTool{
				Type:     "function",
				Function: *function,
			})
		}
	} else if request.Tools != nil {
		zhipuRequest.Tools = make([]ZhipuTool, 0, len(request.Tools))
		for _, tool := range request.Tools {
			zhipuRequest.Tools = append(zhipuRequest.Tools, ZhipuTool{
				Type:     "function",
				Function: tool.Function,
			})
		}
	}

	return zhipuRequest
}

// 转换为OpenAI聊天流式请求体
func (h *zhipuStreamHandler) handlerStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data: 或者 meta:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	*rawLine = (*rawLine)[6:]

	if strings.HasPrefix(string(*rawLine), "[DONE]") {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	zhipuResponse := &ZhipuStreamResponse{}
	err := json.Unmarshal(*rawLine, zhipuResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	error := errorHandle(&zhipuResponse.Error)
	if error != nil {
		errChan <- error
		return
	}

	h.convertToOpenaiStream(zhipuResponse, dataChan, errChan)
}

func (h *zhipuStreamHandler) convertToOpenaiStream(zhipuResponse *ZhipuStreamResponse, dataChan chan string, errChan chan error) {
	streamResponse := types.ChatCompletionStreamResponse{
		ID:      zhipuResponse.ID,
		Object:  "chat.completion.chunk",
		Created: zhipuResponse.Created,
		Model:   h.Request.Model,
	}

	choice := zhipuResponse.Choices[0]

	if choice.Delta.ToolCalls != nil {
		choices := choice.ConvertOpenaiStream()
		for _, choice := range choices {
			chatCompletionCopy := streamResponse
			chatCompletionCopy.Choices = []types.ChatCompletionStreamChoice{choice}
			responseBody, _ := json.Marshal(chatCompletionCopy)
			dataChan <- string(responseBody)
		}
	} else {
		streamResponse.Choices = []types.ChatCompletionStreamChoice{choice}
		responseBody, _ := json.Marshal(streamResponse)
		dataChan <- string(responseBody)
		time.Sleep(20 * time.Millisecond)
	}

	if zhipuResponse.Usage != nil {
		*h.Usage = *zhipuResponse.Usage
	}
}
