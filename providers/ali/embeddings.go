package ali

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

// 嵌入请求处理
func (aliResponse *AliEmbeddingResponse) ResponseHandler(resp *http.Response) (any, *types.OpenAIErrorWithStatusCode) {
	if aliResponse.Code != "" {
		return nil, &types.OpenAIErrorWithStatusCode{
			OpenAIError: types.OpenAIError{
				Message: aliResponse.Message,
				Type:    aliResponse.Code,
				Param:   aliResponse.RequestId,
				Code:    aliResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}
	}

	openAIEmbeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data:   make([]types.Embedding, 0, len(aliResponse.Output.Embeddings)),
		Model:  "text-embedding-v1",
		Usage:  &types.Usage{TotalTokens: aliResponse.Usage.TotalTokens},
	}

	for _, item := range aliResponse.Output.Embeddings {
		openAIEmbeddingResponse.Data = append(openAIEmbeddingResponse.Data, types.Embedding{
			Object:    `embedding`,
			Index:     item.TextIndex,
			Embedding: item.Embedding,
		})
	}

	return openAIEmbeddingResponse, nil
}

// 获取嵌入请求体
func (p *AliProvider) getEmbeddingsRequestBody(request *types.EmbeddingRequest) *AliEmbeddingRequest {
	return &AliEmbeddingRequest{
		Model: "text-embedding-v1",
		Input: struct {
			Texts []string `json:"texts"`
		}{
			Texts: request.ParseInput(),
		},
	}
}

func (p *AliProvider) EmbeddingsAction(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody := p.getEmbeddingsRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.Embeddings, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	aliEmbeddingResponse := &AliEmbeddingResponse{}
	errWithCode = p.SendRequest(req, aliEmbeddingResponse)
	if errWithCode != nil {
		return
	}
	usage = &types.Usage{TotalTokens: aliEmbeddingResponse.Usage.TotalTokens}

	return usage, nil
}
