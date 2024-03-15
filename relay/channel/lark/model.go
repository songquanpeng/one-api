package lark

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Input struct {
	//Prompt   string       `json:"prompt"`
	Messages []Message `json:"messages"`
}

type Parameters struct {
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	Seed              uint64  `json:"seed,omitempty"`
	EnableSearch      bool    `json:"enable_search,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
	MaxTokens         int     `json:"max_tokens,omitempty"`
	Temperature       float64 `json:"temperature,omitempty"`
}

type Model struct {
	Name string `json:"name"`
}

type ChatRequest struct {
	Model      Model      `json:"model"`
	Messages   []Message  `json:"messages"`
	Parameters Parameters `json:"parameters,omitempty"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type Embedding struct {
	Embedding []float64 `json:"embedding"`
	TextIndex int       `json:"text_index"`
}

type EmbeddingResponse struct {
	Output struct {
		Embeddings []Embedding `json:"embeddings"`
	} `json:"output"`
	Usage Usage `json:"usage"`
	Error
}

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type Usage struct {
	TotalTokens      int `json:"total_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type Output struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type ChatResponse struct {
	RequestId string `json:"request_id"`
	Choice    Choice `json:"choice"`
	Usage     Usage  `json:"usage"`
	Error
}
