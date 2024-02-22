package gemini

type ChatRequest struct {
	Contents         []ChatContent        `json:"contents"`
	SafetySettings   []ChatSafetySettings `json:"safety_settings,omitempty"`
	GenerationConfig ChatGenerationConfig `json:"generation_config,omitempty"`
	Tools            []ChatTools          `json:"tools,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type ChatContent struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

type ChatSafetySettings struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type ChatTools struct {
	FunctionDeclarations any `json:"functionDeclarations,omitempty"`
}

type ChatGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            float64  `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	CandidateCount  int      `json:"candidateCount,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}
