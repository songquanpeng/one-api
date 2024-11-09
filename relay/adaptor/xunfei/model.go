package xunfei

import (
	"github.com/songquanpeng/one-api/relay/model"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Functions struct {
	Text []model.Function `json:"text,omitempty"`
}

type ChatRequest struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Parameter struct {
		Chat struct {
			Domain      string   `json:"domain,omitempty"`
			Temperature *float64 `json:"temperature,omitempty"`
			TopK        int      `json:"top_k,omitempty"`
			MaxTokens   int      `json:"max_tokens,omitempty"`
			Auditing    bool     `json:"auditing,omitempty"`
		} `json:"chat"`
	} `json:"parameter"`
	Payload struct {
		Message struct {
			Text []Message `json:"text"`
		} `json:"message"`
		Functions *Functions `json:"functions,omitempty"`
	} `json:"payload"`
}

type ChatResponseTextItem struct {
	Content      string          `json:"content"`
	Role         string          `json:"role"`
	Index        int             `json:"index"`
	ContentType  string          `json:"content_type"`
	FunctionCall *model.Function `json:"function_call"`
}

type ChatResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int                    `json:"status"`
			Seq    int                    `json:"seq"`
			Text   []ChatResponseTextItem `json:"text"`
		} `json:"choices"`
		Usage struct {
			//Text struct {
			//	QuestionTokens   string `json:"question_tokens"`
			//	PromptTokens     string `json:"prompt_tokens"`
			//	CompletionTokens string `json:"completion_tokens"`
			//	TotalTokens      string `json:"total_tokens"`
			//} `json:"text"`
			Text model.Usage `json:"text"`
		} `json:"usage"`
	} `json:"payload"`
}
