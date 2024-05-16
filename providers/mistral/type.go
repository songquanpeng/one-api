package mistral

import (
	"encoding/json"
	"one-api/types"
)

type MistralChatCompletionRequest struct {
	Model       string                        `json:"model" binding:"required"`
	Messages    []types.ChatCompletionMessage `json:"messages" binding:"required"`
	Temperature float64                       `json:"temperature,omitempty"` // 0-1
	MaxTokens   int                           `json:"max_tokens,omitempty"`
	TopP        float64                       `json:"top_p,omitempty"` // 0-1
	N           int                           `json:"n,omitempty"`
	Stream      bool                          `json:"stream,omitempty"`
	Tools       []*types.ChatCompletionTool   `json:"tools,omitempty"`
	ToolChoice  string                        `json:"tool_choice,omitempty"`
	Seed        *int                          `json:"seed,omitempty"`
	SafePrompt  bool                          `json:"safe_prompt,omitempty"`
}

type MistralError struct {
	Object  string               `json:"object"`
	Type    string               `json:"type,omitempty"`
	Message MistralErrorMessages `json:"message,omitempty"`
}

type MistralErrorMessages struct {
	Detail []MistralErrorDetail `json:"detail,omitempty"`
}

type MistralErrorDetail struct {
	Type  string `json:"type"`
	Loc   any    `json:"loc"`
	Msg   string `json:"msg"`
	Input string `json:"input"`
	Ctx   any    `json:"ctx"`
}

func (m *MistralErrorDetail) errorMsg() string {
	// 循环Loc，拼接成字符串
	// 返回字符串
	var errMsg string

	locStr, _ := json.Marshal(m.Loc)

	errMsg += "Loc:" + string(locStr) + "Msg:" + m.Msg

	return errMsg
}

type ChatCompletionStreamResponse struct {
	types.ChatCompletionStreamResponse
	Usage *types.Usage `json:"usage,omitempty"`
}

type ModelListResponse struct {
	Object string         `json:"object"`
	Data   []ModelDetails `json:"data"`
}

type ModelDetails struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}
