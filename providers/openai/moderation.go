package openai

import (
	"net/http"
	"one-api/common/config"
	"one-api/types"
)

func (p *OpenAIProvider) CreateModeration(request *types.ModerationRequest) (*types.ModerationResponse, *types.OpenAIErrorWithStatusCode) {

	req, errWithCode := p.GetRequestTextBody(config.RelayModeModerations, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &OpenAIProviderModerationResponse{}
	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, response, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	openaiErr := ErrorHandle(&response.OpenAIErrorResponse)
	if openaiErr != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *openaiErr,
			StatusCode:  http.StatusBadRequest,
		}
		return nil, errWithCode
	}

	p.Usage.TotalTokens = p.Usage.PromptTokens

	return &response.ModerationResponse, nil
}
