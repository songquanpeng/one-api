package vertexai

import "github.com/songquanpeng/one-api/relay/adaptor/anthropic"

type Request struct {
	// AnthropicVersion must be "vertex-2023-10-16"
	AnthropicVersion string `json:"anthropic_version"`
	// Model            string              `json:"model"`
	Messages      []anthropic.Message `json:"messages"`
	System        string              `json:"system,omitempty"`
	MaxTokens     int                 `json:"max_tokens,omitempty"`
	StopSequences []string            `json:"stop_sequences,omitempty"`
	Stream        bool                `json:"stream,omitempty"`
	Temperature   float64             `json:"temperature,omitempty"`
	TopP          float64             `json:"top_p,omitempty"`
	TopK          int                 `json:"top_k,omitempty"`
	Tools         []anthropic.Tool    `json:"tools,omitempty"`
	ToolChoice    any                 `json:"tool_choice,omitempty"`
}
