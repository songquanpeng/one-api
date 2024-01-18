package openai

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (p *OpenAIProvider) CreateImageVariations(request *types.ImageEditRequest) (*types.ImageResponse, *types.OpenAIErrorWithStatusCode) {
	req, errWithCode := p.getRequestImageBody(common.RelayModeImagesVariations, request.Model, request)
	if errWithCode != nil {
		return nil, errWithCode
	}
	defer req.Body.Close()

	response := &OpenAIProviderImageResponse{}
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

	return &response.ImageResponse, nil
}
