package providers

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

type OpenAIProviderEmbeddingsResponse struct {
	types.EmbeddingResponse
	types.OpenAIErrorResponse
}

func (c *OpenAIProviderEmbeddingsResponse) requestHandler(resp *http.Response) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		openAIErrorWithStatusCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil
}

func (p *OpenAIProvider) EmbeddingsResponse(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {

	requestBody, err := p.getRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, types.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.Embeddings, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	openAIProviderEmbeddingsResponse := &OpenAIProviderEmbeddingsResponse{}
	openAIErrorWithStatusCode = p.sendRequest(req, openAIProviderEmbeddingsResponse)
	if openAIErrorWithStatusCode != nil {
		return
	}

	usage = openAIProviderEmbeddingsResponse.Usage

	return
}
