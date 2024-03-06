package claude

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ClaudeMetadata struct {
	UserId string `json:"user_id"`
}

type ResContent struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeRequest struct {
	Model         string    `json:"model"`
	System        string    `json:"system,omitempty"`
	Messages      []Message `json:"messages"`
	MaxTokens     int       `json:"max_tokens"`
	StopSequences []string  `json:"stop_sequences,omitempty"`
	Temperature   float64   `json:"temperature,omitempty"`
	TopP          float64   `json:"top_p,omitempty"`
	TopK          int       `json:"top_k,omitempty"`
	//ClaudeMetadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens,omitempty"`
}
type ClaudeResponse struct {
	Content      []ResContent `json:"content"`
	Id           string       `json:"id"`
	Role         string       `json:"role"`
	StopReason   string       `json:"stop_reason"`
	StopSequence string       `json:"stop_sequence,omitempty"`
	Model        string       `json:"model"`
	Usage        `json:"usage,omitempty"`
	Error        ClaudeError `json:"error,omitempty"`
}

type Delta struct {
	Type         string `json:"type,omitempty"`
	Text         string `json:"text,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

type ClaudeStreamResponse struct {
	Type    string         `json:"type"`
	Message ClaudeResponse `json:"message,omitempty"`
	Index   int            `json:"index,omitempty"`
	Delta   Delta          `json:"delta,omitempty"`
	Usage   Usage          `json:"usage,omitempty"`
	Error   ClaudeError    `json:"error,omitempty"`
}
