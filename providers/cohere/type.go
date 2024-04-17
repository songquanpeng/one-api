package cohere

import "one-api/types"

type ChatHistory struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type CohereConnector struct {
	ID                string `json:"id"`
	UserAccessToken   string `json:"user_access_token,omitempty"`
	ContinueOnFailure bool   `json:"continue_on_failure,omitempty"`
	Options           any    `json:"options,omitempty"`
}

type CohereRequest struct {
	Message          string                          `json:"message"`
	Model            string                          `json:"model,omitempty"`
	Stream           bool                            `json:"stream,omitempty"`
	Preamble         string                          `json:"preamble,omitempty"`
	ChatHistory      []ChatHistory                   `json:"chat_history,omitempty"`
	ConversationId   string                          `json:"conversation_id,omitempty"`
	PromptTruncation string                          `json:"prompt_truncation,omitempty"`
	Connectors       []CohereConnector               `json:"connectors,omitempty"`
	Temperature      float64                         `json:"temperature,omitempty"`
	MaxTokens        int                             `json:"max_tokens,omitempty"`
	MaxInputTokens   int                             `json:"max_input_tokens,omitempty"`
	K                int                             `json:"k,omitempty"`
	P                float64                         `json:"p,omitempty"`
	Seed             *int                            `json:"seed,omitempty"`
	StopSequences    any                             `json:"stop_sequences,omitempty"`
	FrequencyPenalty float64                         `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64                         `json:"presence_penalty,omitempty"`
	Tools            []*types.ChatCompletionFunction `json:"tools,omitempty"`
	ToolResults      any                             `json:"tool_results,omitempty"`
	// SearchQueriesOnly bool              `json:"search_queries_only,omitempty"`
}

type APIVersion struct {
	Version string `json:"version"`
}

type Tokens struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	SearchUnits     int `json:"search_units,omitempty"`
	Classifications int `json:"classifications,omitempty"`
}

type Meta struct {
	APIVersion  APIVersion `json:"api_version"`
	BilledUnits Tokens     `json:"billed_units"`
	Tokens      Tokens     `json:"tokens"`
}

type CohereToolCall struct {
	Name       string `json:"name,omitempty"`
	Parameters any    `json:"parameters,omitempty"`
}

type CohereResponse struct {
	Text         string           `json:"text,omitempty"`
	ResponseID   string           `json:"response_id,omitempty"`
	GenerationID string           `json:"generation_id,omitempty"`
	ChatHistory  []ChatHistory    `json:"chat_history,omitempty"`
	FinishReason string           `json:"finish_reason,omitempty"`
	ToolCalls    []CohereToolCall `json:"tool_calls,omitempty"`
	Meta         Meta             `json:"meta,omitempty"`
	CohereError
}

type CohereError struct {
	Message string `json:"message,omitempty"`
}

type CohereStreamResponse struct {
	IsFinished   bool             `json:"is_finished"`
	EventType    string           `json:"event_type"`
	GenerationID string           `json:"generation_id,omitempty"`
	Text         string           `json:"text,omitempty"`
	Response     CohereResponse   `json:"response,omitempty"`
	FinishReason string           `json:"finish_reason,omitempty"`
	ToolCalls    []CohereToolCall `json:"tool_calls,omitempty"`
}
