package minimax

import "one-api/types"

type MiniMaxChatRequest struct {
	Model            string                          `json:"model"`
	Stream           bool                            `json:"stream,omitempty"`
	TokensToGenerate int                             `json:"tokens_to_generate,omitempty"`
	Temperature      float64                         `json:"temperature,omitempty"`
	TopP             float64                         `json:"top_p,omitempty"`
	Messages         []MiniMaxChatMessage            `json:"messages"`
	BotSetting       []MiniMaxBotSetting             `json:"bot_setting,omitempty"`
	ReplyConstraints ReplyConstraints                `json:"reply_constraints,omitempty"`
	Functions        []*types.ChatCompletionFunction `json:"functions,omitempty"`
}

type MiniMaxChatMessage struct {
	SenderType   string                                 `json:"sender_type"`
	SenderName   string                                 `json:"sender_name"`
	Text         string                                 `json:"text"`
	FunctionCall *types.ChatCompletionToolCallsFunction `json:"function_call,omitempty"`
}

type MiniMaxBotSetting struct {
	BotName string `json:"bot_name"`
	Content string `json:"content"`
}

type ReplyConstraints struct {
	SenderType string `json:"sender_type"`
	SenderName string `json:"sender_name"`
}

type MiniMaxChatResponse struct {
	Created             int64                                  `json:"created"`
	Model               string                                 `json:"model"`
	Reply               string                                 `json:"reply"`
	InputSensitive      bool                                   `json:"input_sensitive,omitempty"`
	InputSensitiveType  int64                                  `json:"input_sensitive_type,omitempty"`
	OutputSensitive     bool                                   `json:"output_sensitive"`
	OutputSensitiveType int64                                  `json:"output_sensitive_type,omitempty"`
	Choices             []Choice                               `json:"choices"`
	Usage               *Usage                                 `json:"usage,omitempty"`
	ID                  string                                 `json:"id,omitempty"`
	RequestID           string                                 `json:"request_id"`
	FunctionCall        *types.ChatCompletionToolCallsFunction `json:"function_call,omitempty"`
	MiniMaxBaseResp
}

type Choice struct {
	Messages     []MiniMaxChatMessage `json:"messages"`
	Index        int                  `json:"index"`
	FinishReason string               `json:"finish_reason"`
}

type Usage struct {
	TotalTokens int `json:"total_tokens"`
}

type MiniMaxBaseResp struct {
	BaseResp BaseResp `json:"base_resp"`
}

type BaseResp struct {
	StatusCode int64  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

type MiniMaxEmbeddingRequest struct {
	Model string   `json:"model"`
	Texts []string `json:"texts"`
	Type  string   `json:"type"`
}

type MiniMaxEmbeddingResponse struct {
	Vectors     []any `json:"vectors"`
	TotalTokens int   `json:"total_tokens"`
	MiniMaxBaseResp
}
