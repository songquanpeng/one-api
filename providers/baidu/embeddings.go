package baidu

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (p *BaiduProvider) getEmbeddingsRequestBody(request *types.EmbeddingRequest) *BaiduEmbeddingRequest {
	return &BaiduEmbeddingRequest{
		Input: request.ParseInput(),
	}
}

func (baiduResponse *BaiduEmbeddingResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if baiduResponse.ErrorMsg != "" {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: baiduResponse.ErrorMsg,
				Type:    "baidu_error",
				Param:   "",
				Code:    baiduResponse.ErrorCode,
			},
			StatusCode: resp.StatusCode,
		}
	}

	openAIEmbeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data:   make([]types.Embedding, 0, len(baiduResponse.Data)),
		Model:  "text-embedding-v1",
		Usage:  &baiduResponse.Usage,
	}

	for _, item := range baiduResponse.Data {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, types.Embedding{
			Object:    item.Object,
			Index:     item.Index,
			Embedding: item.Embedding,
		})
	}

	return openAIEmbeddingResponse, nil
}

func (p *BaiduProvider) EmbeddingsAction(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody := p.getEmbeddingsRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.Embeddings, request.Model)
	if fullRequestURL == "" {
		return nil, types.ErrorWrapper(nil, "invalid_baidu_config", http.StatusInternalServerError)
	}

	headers := p.GetRequestHeaders()
	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	baiduEmbeddingResponse := &BaiduEmbeddingResponse{}
	errWithCode = p.SendRequest(req, baiduEmbeddingResponse, false)
	if errWithCode != nil {
		return
	}
	usage = &baiduEmbeddingResponse.Usage

	return usage, nil
}
