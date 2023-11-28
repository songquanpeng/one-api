package providers

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

type AliEmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type AliEmbedding struct {
	Embedding []float64 `json:"embedding"`
	TextIndex int       `json:"text_index"`
}

type AliEmbeddingResponse struct {
	Output struct {
		Embeddings []AliEmbedding `json:"embeddings"`
	} `json:"output"`
	Usage AliUsage `json:"usage"`
	AliError
}

func (aliResponse *AliEmbeddingResponse) requestHandler(resp *http.Response) (OpenAIResponse any, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
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

func (p *AliAIProvider) getEmbeddingsRequestBody(request *types.EmbeddingRequest) *AliEmbeddingRequest {
	return &AliEmbeddingRequest{
		Model: "text-embedding-v1",
		Input: struct {
			Texts []string `json:"texts"`
		}{
			Texts: request.ParseInput(),
		},
	}
}

func (p *AliAIProvider) EmbeddingsResponse(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {

	requestBody := p.getEmbeddingsRequestBody(request)
	fullRequestURL := p.GetFullRequestURL(p.Embeddings, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	aliEmbeddingResponse := &AliEmbeddingResponse{}
	openAIErrorWithStatusCode = p.sendRequest(req, aliEmbeddingResponse)
	if openAIErrorWithStatusCode != nil {
		return
	}
	usage = &types.Usage{TotalTokens: aliEmbeddingResponse.Usage.TotalTokens}

	return usage, nil
}
