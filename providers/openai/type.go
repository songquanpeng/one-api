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
