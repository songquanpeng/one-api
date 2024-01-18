package ali

import (
	"one-api/types"
)

type AliError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type AliUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type AliMessage struct {
	Content any    `json:"content"`
	Role    string `json:"role"`
}

type AliMessagePart struct {
	Text  string `json:"text,omitempty"`
	Image string `json:"image,omitempty"`
}

type AliInput struct {
	// Prompt  string       `json:"prompt"`
	Messages []AliMessage `json:"messages"`
}

type AliParameters struct {
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	Seed              uint64  `json:"seed,omitempty"`
	EnableSearch      bool    `json:"enable_search,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
	ResultFormat      string  `json:"result_format,omitempty"`
}

type AliChatRequest struct {
	Model      string        `json:"model"`
	Input      AliInput      `json:"input"`
	Parameters AliParameters `json:"parameters,omitempty"`
}

type AliChoice struct {
	FinishReason string                      `json:"finish_reason"`
	Message      types.ChatCompletionMessage `json:"message"`
}

type AliOutput struct {
	Choices      []types.ChatCompletionChoice `json:"choices"`
	FinishReason string                       `json:"finish_reason,omitempty"`
}

func (o *AliOutput) ToChatCompletionChoices() []types.ChatCompletionChoice {
	for i := range o.Choices {
		_, ok := o.Choices[i].Message.Content.(string)
		if ok {
			continue
		}

		o.Choices[i].Message.Content = o.Choices[i].Message.ParseContent()
	}
	return o.Choices
}

type AliChatResponse struct {
	Output AliOutput `json:"output"`
	Usage  AliUsage  `json:"usage"`
	AliError
}

type AliEmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Texts []string `json:"texts"`
	} `json:"input"`
	Parameters *struct {
		TextType string `json:"text_type,omitempty"`
	} `json:"parameters,omitempty"`
}

type AliEmbedding struct {
	Embedding []float64 `json:"embedding"`
	TextIndex int       `json:"text_index"`
}

type AliEmbeddingResponse struct {
	Output struct {
		Embeddings []AliEmbedding `json:"embeddings"`
	} `json:"output"`
	Usage AliUsage `json:"usage"`
	AliError
}
