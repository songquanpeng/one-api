package cloudflareAI

import "one-api/types"

type CloudflareAIError struct {
	Error []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors,omitempty"`
	Success bool `json:"success"`
}

type ImageRequest struct {
	Prompt   string      `json:"prompt"`
	Image    interface{} `json:"image,omitempty"` // 可以是 string 或者 ImageObject
	Mask     interface{} `json:"mask,omitempty"`  // 可以是 string 或者 MaskObject
	NumSteps int         `json:"num_steps,omitempty"`
	Strength float64     `json:"strength,omitempty"`
	Guidance float64     `json:"guidance,omitempty"`
}

type ImageObject struct {
	Image []float64 `json:"image"`
}

type MaskObject struct {
	Mask []float64 `json:"mask"`
}

type ChatRequest struct {
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRespone struct {
	Result ChatResult `json:"result,omitempty"`
	CloudflareAIError
}

type ChatResult struct {
	Response string `json:"response"`
}

type AudioResponse struct {
	Result AudioResult `json:"result,omitempty"`
	CloudflareAIError
}

type AudioResult struct {
	Text      string                 `json:"text,omitempty"`
	WordCount int                    `json:"word_count,omitempty"`
	Words     []types.AudioWordsList `json:"words,omitempty"`
	Vtt       string                 `json:"vtt,omitempty"`
}
