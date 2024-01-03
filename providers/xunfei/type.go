package xunfei

import "one-api/types"

type XunfeiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type XunfeiChatPayloadMessage struct {
	Text []XunfeiMessage `json:"text"`
}

type XunfeiChatPayloadFunctions struct {
	Text []*types.ChatCompletionFunction `json:"text"`
}

type XunfeiChatPayload struct {
	Message   XunfeiChatPayloadMessage    `json:"message"`
	Functions *XunfeiChatPayloadFunctions `json:"functions,omitempty"`
}

type XunfeiParameterChat struct {
	Domain      string  `json:"domain,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Auditing    bool    `json:"auditing,omitempty"`
}

type XunfeiChatRequestParameter struct {
	Chat XunfeiParameterChat `json:"chat"`
}

type XunfeiChatRequest struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Parameter XunfeiChatRequestParameter `json:"parameter"`
	Payload   XunfeiChatPayload          `json:"payload"`
}

type XunfeiChatResponseTextItem struct {
	Content      string                                 `json:"content"`
	Role         string                                 `json:"role"`
	Index        int                                    `json:"index"`
	ContentType  string                                 `json:"content_type,omitempty"`
	FunctionCall *types.ChatCompletionToolCallsFunction `json:"function_call,omitempty"`
}

type XunfeiChatResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int                          `json:"status"`
			Seq    int                          `json:"seq"`
			Text   []XunfeiChatResponseTextItem `json:"text"`
		} `json:"choices"`
		Usage struct {
			Text types.Usage `json:"text"`
		} `json:"usage"`
	} `json:"payload"`
}
