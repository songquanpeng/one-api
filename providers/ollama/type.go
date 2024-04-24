package ollama

import "time"

type OllamaError struct {
	Error string `json:"error,omitempty"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages,omitempty"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format,omitempty"`
	Options  Option    `json:"options,omitempty"`
}

type Option struct {
	Temperature float64 `json:"temperature,omitempty"`
	Seed        *int    `json:"seed,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
}

type ChatResponse struct {
	OllamaError
	Model           string    `json:"model"`
	CreatedAt       time.Time `json:"created_at"`
	Message         Message   `json:"message,omitempty"`
	Done            bool      `json:"done"`
	EvalCount       int       `json:"eval_count,omitempty"`
	PromptEvalCount int       `json:"prompt_eval_count,omitempty"`
}

type Message struct {
	Role    string   `json:"role,omitempty"`
	Content string   `json:"content,omitempty"`
	Images  []string `json:"images,omitempty"`
}

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	OllamaError
	Embedding []float64 `json:"embedding,omitempty"`
}
