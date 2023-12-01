package openai

import "one-api/types"

type OpenAIProviderChatResponse struct {
	types.ChatCompletionResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderChatStreamResponse struct {
	types.ChatCompletionStreamResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderCompletionResponse struct {
	types.CompletionResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderEmbeddingsResponse struct {
	types.EmbeddingResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderModerationResponse struct {
	types.ModerationResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderTranscriptionsResponse struct {
	types.AudioResponse
	types.OpenAIErrorResponse
}

type OpenAIProviderTranscriptionsTextResponse string

func (a *OpenAIProviderTranscriptionsTextResponse) GetString() *string {
	return (*string)(a)
}

type OpenAIProviderImageResponseResponse struct {
	types.ImageResponse
	types.OpenAIErrorResponse
}
