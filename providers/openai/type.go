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

type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
	AccessUntil        int64   `json:"access_until"`
}

type OpenAIUsageResponse struct {
	Object string `json:"object"`
	//DailyCosts []OpenAIUsageDailyCost `json:"daily_costs"`
	TotalUsage float64 `json:"total_usage"` // unit: 0.01 dollar
}
