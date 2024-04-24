package vertex

type Request struct {
	AnthropicVersion string    `json:"anthropic_version"`
	Messages         []Message `json:"messages"`
	MaxTokens        int       `json:"max_tokens"`
	Stream           bool      `json:"stream"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type   string  `json:"type"`
	Source *Source `json:"source,omitempty"`
	Text   string  `json:"text,omitempty"`
}

type Source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Response struct {
	Id           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Model        string    `json:"model"`
	StopReason   *string   `json:"stop_reason"`
	StopSequence *string   `json:"stop_sequence"`
	Usage        Usage     `json:"usage"`
	Error        Error     `json:"error"`
}

type Delta struct {
	Type         string  `json:"type"`
	Text         string  `json:"text"`
	StopReason   *string `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
}

type StreamResponse struct {
	Type         string    `json:"type"`
	Message      *Response `json:"message"`
	Index        int       `json:"index"`
	ContentBlock *Content  `json:"content_block"`
	Delta        *Delta    `json:"delta"`
	Usage        *Usage    `json:"usage"`
}
