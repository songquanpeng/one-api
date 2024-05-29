package baidu

import (
	"net/http"
	"one-api/common"
	"one-api/common/config"
	"one-api/types"
)

func (p *BaiduProvider) CreateEmbeddings(request *types.EmbeddingRequest) (*types.EmbeddingResponse, *types.OpenAIErrorWithStatusCode) {
	url, errWithCode := p.GetSupportedAPIUri(config.RelayModeEmbeddings)
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

	aliRequest := convertFromEmbeddingOpenai(request)
	// 创建请求
	req, err := p.Requester.NewRequest(http.MethodPost, fullRequestURL, p.Requester.WithBody(aliRequest), p.Requester.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}
	defer req.Body.Close()

	baiduResponse := &BaiduEmbeddingResponse{}

	// 发送请求
	_, errWithCode = p.Requester.SendRequest(req, baiduResponse, false)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.convertToEmbeddingOpenai(baiduResponse, request)
}

func convertFromEmbeddingOpenai(request *types.EmbeddingRequest) *BaiduEmbeddingRequest {
	return &BaiduEmbeddingRequest{
		Input: request.ParseInput(),
	}
}

func (p *BaiduProvider) convertToEmbeddingOpenai(response *BaiduEmbeddingResponse, request *types.EmbeddingRequest) (openaiResponse *types.EmbeddingResponse, errWithCode *types.OpenAIErrorWithStatusCode) {
	aiError := errorHandle(&response.BaiduError)
	if aiError != nil {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: *aiError,
			StatusCode:  http.StatusBadRequest,
		}
		return
	}

	openAIEmbeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data:   make([]types.Embedding, 0, len(response.Data)),
		Model:  request.Model,
		Usage:  &response.Usage,
	}

	for _, item := range response.Data {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, types.Embedding{
			Object:    item.Object,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}

	*p.Usage = response.Usage

	return openAIEmbeddingResponse, nil
}
