package zhipu

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *ZhipuProvider) CreateEmbeddings(request *types.EmbeddingRequest) (*types.EmbeddingResponse, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeEmbeddings)
	if errWithCode != nil {
		return nil, errWithCode
	}
	// 获取请求地址
	fullRequestURL := p.GetFullRequestURL(url)
	if fullRequestURL == "" {
		return nil, common.ErrorWrapper(nil, "invalid_zhipu_config", http.StatusInternalServerError)
	}

	// 获取请求头
	headers := p.GetRequestHeaders()

	aliRequest := convertFromEmbeddingOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(aliRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	defer req.Body.Close()

	zhipuResponse := &ZhipuEmbeddingResponse{}

	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, zhipuResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToEmbeddingOpenai(zhipuResponse, request)
}

func convertFromEmbeddingOpenai(request *types.EmbeddingRequest) *ZhipuEmbeddingRequest {
	return &ZhipuEmbeddingRequest{
		Model: request.Model,
		Input: request.ParseInputString(),
	}
}

func (p *ZhipuProvider) convertToEmbeddingOpenai(response *ZhipuEmbeddingResponse, request *types.EmbeddingRequest) (openaiResponse *types.EmbeddingResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.Error)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openAIEmbeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data:   response.Data,
		Model:  request.Model,
		Usage:  response.Usage,
	}

	*p.Usage = *response.Usage

	return openAIEmbeddingResponse, nil
}
