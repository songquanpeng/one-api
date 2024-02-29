package model

type ResponseFormat struct {
	Type string `json:"type,omitempty"`
}

type GeneralOpenAIRequest struct {
	Model            string          `json:"model,omitempty"`
	Messages         []Message       `json:"messages,omitempty"`
	Prompt           any             `json:"prompt,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	MaxTokens        int             `json:"max_tokens,omitempty"`
	Temperature      float64         `json:"temperature,omitempty"`
	TopP             float64         `json:"top_p,omitempty"`
	N                int             `json:"n,omitempty"`
	Input            any             `json:"input,omitempty"`
	Instruction      string          `json:"instruction,omitempty"`
	Size             string          `json:"size,omitempty"`
	Functions        any             `json:"functions,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	Seed             float64         `json:"seed,omitempty"`
	Tools            any             `json:"tools,omitempty"`
	ToolChoice       any             `json:"tool_choice,omitempty"`
	User             string          `json:"user,omitempty"`
}

func (r GeneralOpenAIRequest) ParseInput() []string {
	if r.Input == nil {
		return nil
	}
	var input []string
	switch r.Input.(type) {
	case string:
		input = []string{r.Input.(string)}
	case []any:
		input = make([]string, 0, len(r.Input.([]any)))
		for _, item := range r.Input.([]any) {
			if str, ok := item.(string); ok {
				input = append(input, str)
			}
		}
	}
	return input
}
