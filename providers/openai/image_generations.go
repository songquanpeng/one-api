package openai

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (c *OpenAIProviderImageResponseResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil, nil
}

func (p *OpenAIProvider) ImageGenerationsAction(request *types.ImageRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {

	if isWithinRange(request.Model, request.N) == false {
		return nil, common.StringErrorWrapper("n_not_within_range", "n_not_within_range", http.StatusBadRequest)
	}

	requestBody, err := p.GetRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, common.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.ImagesGenerations, request.Model)
	headers := p.GetRequestHeaders()

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	openAIProviderImageResponseResponse := &OpenAIProviderImageResponseResponse{}
	errWithCode = p.SendRequest(req, openAIProviderImageResponseResponse, true)
	if errWithCode != nil {
		return
	}

	usage = &types.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		TotalTokens:      promptTokens,
	}

	return
}

func isWithinRange(element string, value int) bool {
	if _, ok := common.DalleGenerationImageAmounts[element]; !ok {
		return false
	}
	min := common.DalleGenerationImageAmounts[element][0]
	max := common.DalleGenerationImageAmounts[element][1]

	return value >= min && value <= max
}
