package openai

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

func (c *OpenAIProviderChatResponse) ResponseHandler(resp *http.Response) (OpenAIResponse any, errWithCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		errWithCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil, nil
}

func (c *OpenAIProviderChatStreamResponse) responseStreamHandler() (responseText string) {
	for _, choice := range c.Choices {
		responseText += choice.Delta.Content
	}

	return
}

func (p *OpenAIProvider) ChatAction(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, errWithCode *types.OpenAIErrorWithStatusCode) {
	requestBody, err := p.GetRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, common.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream && headers["Accept"] == "" {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, common.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		openAIProviderChatStreamResponse := &OpenAIProviderChatStreamResponse{}
		var textResponse string
		errWithCode, textResponse = p.sendStreamRequest(req, openAIProviderChatStreamResponse)
		if errWithCode != nil {
			return
		}

		usage = &types.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: common.CountTokenText(textResponse, request.Model),
			TotalTokens:      promptTokens + common.CountTokenText(textResponse, request.Model),
		}

	} else {
		openAIProviderChatResponse := &OpenAIProviderChatResponse{}
		errWithCode = p.SendRequest(req, openAIProviderChatResponse, true)
		if errWithCode != nil {
			return
		}

		usage = openAIProviderChatResponse.Usage

		if usage.TotalTokens == 0 {
			completionTokens := 0
			for _, choice := range openAIProviderChatResponse.Choices {
				completionTokens += common.CountTokenText(choice.Message.StringContent(), openAIProviderChatResponse.Model)
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
