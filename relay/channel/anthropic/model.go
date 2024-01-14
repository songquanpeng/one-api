package anthropic

type Metadata struct {
	UserId string `json:"user_id"`
}

type Request struct {
	Model             string   `json:"model"`
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	//Metadata    `json:"metadata,omitempty"`
	Stream bool `json:"stream,omitempty"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Response struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
	Error      Error  `json:"error"`
}
