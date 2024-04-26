package cloudflare

type Request struct {
	Lora        string  `json:"lora,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Prompt      string  `json:"prompt,omitempty"`
	Raw         bool    `json:"raw,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type Result struct {
	Response string `json:"response"`
}

type Response struct {
	Result   Result   `json:"result"`
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
}

type StreamResponse struct {
	Response string `json:"response"`
}
