package openai

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

type OpenAIStreamHandler struct {
	Usage     *types.Usage
	ModelName string
	isAzure   bool
}

func (p *OpenAIProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.GetRequestTextBody(config.RelayModeChatCompletions, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &OpenAIProviderChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 检测是否错误
	openaiErr := ErrorHandle(&response.OpenAIErrorResponse)
	if openaiErr != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *openaiErr,
			StatusCode:  http.StatusBadRequest,
		}
		return nil, errWithCode
	}

	if response.Usage == nil {
		response.Usage = &types.Usage{
			PromptTokens:     p.Usage.PromptTokens,
			CompletionTokens: 0,
			TotalTokens:      0,
		}
		// 那么需要计算
		response.Usage.CompletionTokens = common.CountTokenText(response.GetContent(), request.Model)
		response.Usage.TotalTokens = response.Usage.PromptTokens + response.Usage.CompletionTokens
	}

	*p.Usage = *response.Usage

	return &response.ChatCompletionResponse, nil
}

func (p *OpenAIProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	streamOptions := request.StreamOptions
	// 如果支持流式返回Usage 则需要更改配置：
	if p.SupportStreamOptions {
		request.StreamOptions = &types.StreamOptions{
			IncludeUsage: true,
		}
	} else {
		// 避免误传导致报错
		request.StreamOptions = nil
	}
	req, errWithCode := p.GetRequestTextBody(config.RelayModeChatCompletions, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 恢复原来的配置
	request.StreamOptions = streamOptions

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := OpenAIStreamHandler{
		Usage:     p.Usage,
		ModelName: request.Model,
		isAzure:   p.IsAzure,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.HandlerChatStream)
}

func (h *OpenAIStreamHandler) HandlerChatStream(rawLine *[]byte, dataChan chan string, errChan chan error) {
	// 如果rawLine 前缀不为data:，则直接返回
	if !strings.HasPrefix(string(*rawLine), "data: ") {
		*rawLine = nil
		return
	}

	// 去除前缀
	*rawLine = (*rawLine)[6:]

	// 如果等于 DONE 则结束
	if string(*rawLine) == "[DONE]" {
		errChan <- io.EOF
		*rawLine = requester.StreamClosed
		return
	}

	var openaiResponse OpenAIProviderChatStreamResponse
	err := json.Unmarshal(*rawLine, &openaiResponse)
	if err != nil {
		errChan <- common.ErrorToOpenAIError(err)
		return
	}

	aiError := ErrorHandle(&openaiResponse.OpenAIErrorResponse)
	if aiError != nil {
		errChan <- aiError
		return
	}

	if len(openaiResponse.Choices) == 0 {
		if openaiResponse.Usage != nil {
			*h.Usage = *openaiResponse.Usage
		}
		*rawLine = nil
		return
	}

	dataChan <- string(*rawLine)

	if len(openaiResponse.Choices) > 0 && openaiResponse.Choices[0].Usage != nil {
		*h.Usage = *openaiResponse.Choices[0].Usage
	} else {
		if h.Usage.TotalTokens == 0 {
			h.Usage.TotalTokens = h.Usage.PromptTokens
		}
		countTokenText := common.CountTokenText(openaiResponse.GetResponseText(), h.ModelName)
		h.Usage.CompletionTokens += countTokenText
		h.Usage.TotalTokens += countTokenText
	}
}
