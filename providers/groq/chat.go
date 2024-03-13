package groq

import (
	"net/http"
	"one-api/common"
	"one-api/common/requester"
	"one-api/providers/openai"
	"one-api/types"
)

func (p *GroqProvider) CreateChatCompletion(request *types.ChatCompletionRequest) (openaiResponse *types.ChatCompletionResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	p.getChatRequestBody(request)

	req, errWithCode := p.GetRequestTextBody(common.RelayModeChatCompletions, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &openai.OpenAIProviderChatResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// 检测是否错误
	openaiErr := openai.ErrorHandle(&response.OpenAIErrorResponse)
	if openaiErr != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *openaiErr,
			StatusCode:  http.StatusBadRequest,
		}
		return nil, errWithCode
	}

	*p.Usage = *response.Usage

	return &response.ChatCompletionResponse, nil
}

func (p *GroqProvider) CreateChatCompletionStream(request *types.ChatCompletionRequest) (requester.StreamReaderInterface[string], *types.OpenAIErrorWithStatusCode) {
	p.getChatRequestBody(request)
	req, errWithCode := p.GetRequestTextBody(common.RelayModeChatCompletions, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	// 发送请求
	resp, errWithCode := p.Requester.SendRequestRaw(req)
	if errWithCode != nil {
		return nil, errWithCode
	}

	chatHandler := openai.OpenAIStreamHandler{
		Usage:     p.Usage,
		ModelName: request.Model,
	}

	return requester.RequestStream[string](p.Requester, resp, chatHandler.HandlerChatStream)
}

// 获取聊天请求体
func (p *GroqProvider) getChatRequestBody(request *types.ChatCompletionRequest) {
	if request.Tools != nil {
		request.Tools = nil
	}

	if request.ToolChoice != nil {
		request.ToolChoice = nil
	}

	if request.ResponseFormat != nil {
		request.ResponseFormat = nil
	}

	if request.N > 1 {
		request.N = 1
	}

}
