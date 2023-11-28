package providers

import (
	"net/http"
	"one-api/common"
	"one-api/types"
)

type OpenAIProviderChatResponse struct {
	types.ChatCompletionResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderChatStreamResponse struct {
	types.ChatCompletionStreamResponse
	types.OpenAIErrorResponse
}

func (c *OpenAIProviderChatResponse) requestHandler(resp *http.Response) (openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	if c.Error.Type != "" {
		openAIErrorWithStatusCode = &types.OpenAIErrorWithStatusCode{
			OpenAIError: c.Error,
			StatusCode:  resp.StatusCode,
		}
		return
	}
	return nil
}

func (c *OpenAIProviderChatStreamResponse) requestStreamHandler() (responseText string) {
	for _, choice := range c.Choices {
		responseText += choice.Delta.Content
	}

	return
}

func (p *OpenAIProvider) ChatCompleteResponse(request *types.ChatCompletionRequest, isModelMapped bool, promptTokens int) (usage *types.Usage, openAIErrorWithStatusCode *types.OpenAIErrorWithStatusCode) {
	requestBody, err := p.getRequestBody(&request, isModelMapped)
	if err != nil {
		return nil, types.ErrorWrapper(err, "json_marshal_failed", http.StatusInternalServerError)
	}

	fullRequestURL := p.GetFullRequestURL(p.ChatCompletions, request.Model)
	headers := p.GetRequestHeaders()
	if request.Stream && headers["Accept"] == "" {
		headers["Accept"] = "text/event-stream"
	}

	client := common.NewClient()
	req, err := client.NewRequest(p.Context.Request.Method, fullRequestURL, common.WithBody(requestBody), common.WithHeader(headers))
	if err != nil {
		return nil, types.ErrorWrapper(err, "new_request_failed", http.StatusInternalServerError)
	}

	if request.Stream {
		openAIProviderChatStreamResponse := &OpenAIProviderChatStreamResponse{}
		var textResponse string
		openAIErrorWithStatusCode, textResponse = p.sendStreamRequest(req, openAIProviderChatStreamResponse)
		if openAIErrorWithStatusCode != nil {
			return
		}

		usage = &types.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: common.CountTokenText(textResponse, request.Model),
			TotalTokens:      promptTokens + common.CountTokenText(textResponse, request.Model),
		}

	} else {
		openAIProviderChatResponse := &OpenAIProviderChatResponse{}
		openAIErrorWithStatusCode = p.sendRequest(req, openAIProviderChatResponse)
		if openAIErrorWithStatusCode != nil {
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
