package minimax

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *MiniMaxProvider) CreateEmbeddings(request *types.EmbeddingRequest) (*types.EmbeddingResponse, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeEmbeddings)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_minimax_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	minimaxRequest := convertFromEmbeddingOpenai(request)

	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(minimaxRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	defer req.Body.Close()

	minimaxResponse := &MiniMaxEmbeddingResponse{}

	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, minimaxResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToEmbeddingOpenai(minimaxResponse, request)
}

func convertFromEmbeddingOpenai(request *types.EmbeddingRequest) *MiniMaxEmbeddingRequest {
	minimaxRequest := &MiniMaxEmbeddingRequest{
		Model: request.Model,
		Type:  "db",
	}

	if input, ok := request.Input.(string); ok {
		minimaxRequest.Texts = []string{input}
	} else if inputs, ok := request.Input.([]any); ok {
		for _, item := range inputs {
			if input, ok := item.(string); ok {
				minimaxRequest.Texts = append(minimaxRequest.Texts, input)
			}
		}
	}

	return minimaxRequest
}

func (p *MiniMaxProvider) convertToEmbeddingOpenai(response *MiniMaxEmbeddingResponse, request *types.EmbeddingRequest) (openaiResponse *types.EmbeddingResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.BaseResp)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openaiResponse = &types.EmbeddingResponse{
		Object: "list",
		Model:  request.Model,
	}

	for _, item := range response.Vectors {
		openaiResponse.Data = append(openaiResponse.Data, types.Embedding{
			Object:    "embedding",
			Embedding: item,
		})
	}

	if response.TotalTokens < p.Usage.PromptTokens {
		p.Usage.PromptTokens = response.TotalTokens
	}
	p.Usage.TotalTokens = response.TotalTokens
	p.Usage.CompletionTokens = response.TotalTokens - p.Usage.PromptTokens

	openaiResponse.Usage = p.Usage

	return
}
