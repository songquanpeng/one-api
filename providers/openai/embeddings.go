package openai

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (c *OpenAIProviderEmbeddingsResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil, nil
}

func (p *OpenAIProvider) EmbeddingsAction(request *types.EmbeddingRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	requestBody, err := p.GetRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, common.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.Embeddings, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	openAIProviderEmbeddingsResponse := &OpenAIProviderEmbeddingsResponse{}
	errWithCode = p.SendRequest(req, openAIProviderEmbeddingsResponse, true)
	if errWithCode != nil {
		return
	}

	usage = openAIProviderEmbeddingsResponse.Usage

	return
}
