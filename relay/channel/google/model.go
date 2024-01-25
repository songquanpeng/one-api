package google

import (
	"one-api/relay/channel/openai"
)

type GeminiChatRequest struct {
	Contents         []GeminiChatContent        `json:"contents"`
	SafetySettings   []GeminiChatSafetySettings `json:"safety_settings,omitempty"`
	GenerationConfig GeminiChatGenerationConfig `json:"generation_config,omitempty"`
	Tools            []GeminiChatTools          `json:"tools,omitempty"`
}

type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
}

type GeminiChatContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiChatSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiChatTools struct {
	FunctionDeclarations any `json:"functionDeclarations,omitempty"`
}

type GeminiChatGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            float64  `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	CandidateCount  int      `json:"candidateCount,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type PaLMChatMessage struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type PaLMFilter struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type PaLMPrompt struct {
	Messages []PaLMChatMessage `json:"messages"`
}

type PaLMChatRequest struct {
	Prompt         PaLMPrompt `json:"prompt"`
	Temperature    float64    `json:"temperature,omitempty"`
	CandidateCount int        `json:"candidateCount,omitempty"`
	TopP           float64    `json:"topP,omitempty"`
	TopK           int        `json:"topK,omitempty"`
}

type PaLMError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type PaLMChatResponse struct {
	Candidates []PaLMChatMessage `json:"candidates"`
	Messages   []openai.Message  `json:"messages"`
	Filters    []PaLMFilter      `json:"filters"`
	Error      PaLMError         `json:"error"`
}
