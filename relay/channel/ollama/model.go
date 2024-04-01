package ollama

type Options struct {
	Seed             int     `json:"seed,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	TopP             float64 `json:"top_p,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
}

type Message struct {
	Role    string   `json:"role,omitempty"`
	Content string   `json:"content,omitempty"`
	Images  []string `json:"images,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Stream   bool      `json:"stream"`
	Options  *Options  `json:"options,omitempty"`
}

type ChatResponse struct {
	Model           string  `json:"model,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	Message         Message `json:"message,omitempty"`
	Response        string  `json:"response,omitempty"` // for stream response
	Done            bool    `json:"done,omitempty"`
	TotalDuration   int     `json:"total_duration,omitempty"`
	LoadDuration    int     `json:"load_duration,omitempty"`
	PromptEvalCount int     `json:"prompt_eval_count,omitempty"`
	EvalCount       int     `json:"eval_count,omitempty"`
	EvalDuration    int     `json:"eval_duration,omitempty"`
	Error           string  `json:"error,omitempty"`
}

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Error     string    `json:"error,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`
}
