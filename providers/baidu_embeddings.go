package providers

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

type BaiduEmbeddingRequest struct {
	Input []string `json:"input"`
}

type BaiduEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type BaiduEmbeddingResponse struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Data    []BaiduEmbeddingData `json:"data"`
	Usage   types.Usage          `json:"usage"`
	BaiduError
}

func (p *BaiduProvider) getEmbeddingsRequestBody(request *types.EmbeddingRequest) *BaiduEmbeddingRequest {
	return &BaiduEmbeddingRequest{
		Input: request.ParseInput(),
	}
}

func (baiduResponse *BaiduEmbeddingResponse) requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
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

func (p *BaiduProvider) EmbeddingsResponse(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {

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
	openAIErrorWithStatusCode = p.sendRequest(req, baiduEmbeddingResponse)
	if openAIErrorWithStatusCode != nil {
		return
	}
	usage = &baiduEmbeddingResponse.Usage

	return usage, nil
}
