package openai

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (c *OpenAIProviderCompletionResponse) responseHandler(resp *http.Response) (errWithCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil
}

func (c *OpenAIProviderCompletionResponse) responseStreamHandler() (responseText string) {
	for _, choice := range c.Choices {
		responseText += choice.Text
	}

	return
}

func (p *OpenAIProvider) CompleteAction(request *types.CompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	requestBody, err := p.getRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, types.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.Completions, request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream && headers["Accept"] == "" {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	openAIProviderCompletionResponse := &OpenAIProviderCompletionResponse{}
	if request.Stream {
		// TODO
		var textResponse string
		errWithCode, textResponse = p.sendStreamRequest(req, openAIProviderCompletionResponse)
		if errWithCode != nil {
			return
		}

		usage = &types.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: common.CountTokenText(textResponse, request.Model),
			TotalTokens:      promptTokens + common.CountTokenText(textResponse, request.Model),
		}

	} else {
		errWithCode = p.sendRequest(req, openAIProviderCompletionResponse)
		if errWithCode != nil {
			return
		}

		usage = openAIProviderCompletionResponse.Usage

		if usage.TotalTokens == 0 {
			completionTokens := 0
			for _, choice := range openAIProviderCompletionResponse.Choices {
				completionTokens += common.CountTokenText(choice.Text, openAIProviderCompletionResponse.Model)
			}
			usage = &types.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      promptTokens + completionTokens,
			}
		}
	}
	return
}
