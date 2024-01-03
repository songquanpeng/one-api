package baidu

import (
	"one-api/types"
	"time"
)

type BaiduAccessToken struct {
	AccessToken      string    `json:"access_token"`
	Error            string    `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ExpiresIn        int64     `json:"expires_in,omitempty"`
	ExpiresAt        time.Time `json:"-"`
}

type BaiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaiduChatRequest struct {
	Messages    []BaiduMessage                  `json:"messages"`
	Functions   []*types.ChatCompletionFunction `json:"functions,omitempty"`
	Temperature float64                         `json:"temperature,omitempty"`
	Stream      bool                            `json:"stream"`
	UserId      string                          `json:"user_id,omitempty"`
}

type BaiduChatResponse struct {
	Id               string                                 `json:"id"`
	Object           string                                 `json:"object"`
	Created          int64                                  `json:"created"`
	Result           string                                 `json:"result"`
	IsTruncated      bool                                   `json:"is_truncated"`
	NeedClearHistory bool                                   `json:"need_clear_history"`
	Usage            *types.Usage                           `json:"usage"`
	Model            string                                 `json:"model,omitempty"`
	FunctionCall     *types.ChatCompletionToolCallsFunction `json:"function_call,omitempty"`
	FunctionCate     string                                 `json:"function_cate,omitempty"`
	BaiduError
}

type BaiduEmbeddingRequest struct {
	Input []string `json:"input"`
}

type BaiduEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type BaiduEmbeddingResponse struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Data    []BaiduEmbeddingData `json:"data"`
	Usage   types.Usage          `json:"usage"`
	BaiduError
}

type BaiduChatStreamResponse struct {
	BaiduChatResponse
	SentenceId int  `json:"sentence_id"`
	IsEnd      bool `json:"is_end"`
}

type BaiduError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}
