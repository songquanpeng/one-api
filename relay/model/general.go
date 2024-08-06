package model

type ResponseFormat struct {
	Type string `json:"type,omitempty"`
}

type GeneralOpenAIRequest struct {
	Messages         []Message       `json:"messages,omitempty"`
	Model            string          `json:"model,omitempty"`
	FrequencyPenalty float64         `json:"frequency_penalty,omitempty"`
	MaxTokens        int             `json:"max_tokens,omitempty"`
	N                int             `json:"n,omitempty"`
	PresencePenalty  float64         `json:"presence_penalty,omitempty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"`
	Seed             float64         `json:"seed,omitempty"`
	Stop             any             `json:"stop,omitempty"`
	Stream           bool            `json:"stream,omitempty"`
	Temperature      float64         `json:"temperature,omitempty"`
	TopP             float64         `json:"top_p,omitempty"`
	TopK             int             `json:"top_k,omitempty"`
	Tools            []Tool          `json:"tools,omitempty"`
	ToolChoice       any             `json:"tool_choice,omitempty"`
	FunctionCall     any             `json:"function_call,omitempty"`
	Functions        any             `json:"functions,omitempty"`
	User             string          `json:"user,omitempty"`
	Prompt           any             `json:"prompt,omitempty"`
	Input            any             `json:"input,omitempty"`
	EncodingFormat   string          `json:"encoding_format,omitempty"`
	Dimensions       int             `json:"dimensions,omitempty"`
	Instruction      string          `json:"instruction,omitempty"`
	Size             string          `json:"size,omitempty"`
	NumCtx           int         	 `json:"num_ctx,omitempty"`
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
