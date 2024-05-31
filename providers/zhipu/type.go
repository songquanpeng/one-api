package zhipu

import (
	"one-api/types"
	"time"
)

type ZhipuWebSearch struct {
	Enable      bool   `json:"enable"`
	SearchQuery string `json:"search_query,omitempty"`
}

type ZhipuRetrieval struct {
	KnowledgeId    string `json:"knowledge_id"`
	PromptTemplate string `json:"prompt_template,omitempty"`
}

type ZhipuTool struct {
	Type      string                        `json:"type"`
	Function  *types.ChatCompletionFunction `json:"function,omitempty"`
	WebSearch *ZhipuWebSearch               `json:"web_search,omitempty"`
	Retrieval *ZhipuRetrieval               `json:"retrieval,omitempty"`
}
type ZhipuRequest struct {
	Model       string                        `json:"model"`
	Messages    []types.ChatCompletionMessage `json:"messages"`
	Stream      bool                          `json:"stream,omitempty"`
	Temperature float64                       `json:"temperature,omitempty"`
	TopP        float64                       `json:"top_p,omitempty"`
	MaxTokens   int                           `json:"max_tokens,omitempty"`
	Stop        []string                      `json:"stop,omitempty"`
	Tools       []ZhipuTool                   `json:"tools,omitempty"`
	ToolChoice  any                           `json:"tool_choice,omitempty"`
}

// type ZhipuMessage struct {
// 	Role       string                           `json:"role"`
// 	Content    string                           `json:"content"`
// 	ToolCalls  []*types.ChatCompletionToolCalls `json:"tool_calls,omitempty"`
// 	ToolCallId string                           `json:"tool_call_id,omitempty"`
// }

type ZhipuResponse struct {
	ID      string                       `json:"id"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []types.ChatCompletionChoice `json:"choices"`
	Usage   *types.Usage                 `json:"usage,omitempty"`
	ZhipuResponseError
}

type ZhipuStreamResponse struct {
	ID      string                             `json:"id"`
	Created int64                              `json:"created"`
	Choices []types.ChatCompletionStreamChoice `json:"choices"`
	Usage   *types.Usage                       `json:"usage,omitempty"`
	ZhipuResponseError
}

func (z *ZhipuStreamResponse) GetResponseText() (responseText string) {
	for _, choice := range z.Choices {
		responseText += choice.Delta.Content
	}

	return
}

type ZhipuResponseError struct {
	Error ZhipuError `json:"error,omitempty"`
}

type ZhipuError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ZhipuEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ZhipuEmbeddingResponse struct {
	Model  string            `json:"model"`
	Data   []types.Embedding `json:"data"`
	Object string            `json:"object"`
	Usage  *types.Usage      `json:"usage"`
	ZhipuResponseError
}

type ZhipuImageGenerationRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ZhipuImageGenerationResponse struct {
	Model string                         `json:"model"`
	Data  []types.ImageResponseDataInner `json:"data,omitempty"`
	ZhipuResponseError
}

type zhipuTokenData struct {
	Token      string
	ExpiryTime time.Time
}
