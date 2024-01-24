package zhipu_v4

import (
	"one-api/relay/channel/openai"
	"time"
)

type Message struct {
	Role       string `json:"role,omitempty"`
	Content    string `json:"content,omitempty"`
	ToolCalls  any    `json:"tool_calls,omitempty"`
	ToolCallId any    `json:"tool_call_id,omitempty"`
}

type Request struct {
	Model       string    `json:"model"`
	Stream      bool      `json:"stream,omitempty"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
	RequestId   string    `json:"request_id,omitempty"`
	Tools       any       `json:"tools,omitempty"`
	ToolChoice  any       `json:"tool_choice,omitempty"`
}

type TextResponseChoice struct {
	Index        int `json:"index"`
	Message      `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type Response struct {
	Id                  string               `json:"id"`
	Created             int64                `json:"created"`
	Model               string               `json:"model"`
	TextResponseChoices []TextResponseChoice `json:"choices"`
	openai.Usage        `json:"usage"`
	openai.Error        `json:"error"`
}

type StreamResponseChoice struct {
	Index        int     `json:"index,omitempty"`
	Delta        Message `json:"delta"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

type StreamResponse struct {
	Id           string                 `json:"id"`
	Created      int64                  `json:"created"`
	Choices      []StreamResponseChoice `json:"choices"`
	openai.Usage `json:"usage"`
}

type tokenData struct {
	Token      string
	ExpiryTime time.Time
}
