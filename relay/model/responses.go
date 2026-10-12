package model

// OpenAI Responses API（/v1/responses）数据结构
// https://platform.openai.com/docs/api-reference/responses

// ResponsesRequest Responses API 请求结构
type ResponsesRequest struct {
	Model              string         `json:"model"`
	Input              any            `json:"input,omitempty"`
	Instructions       string         `json:"instructions,omitempty"`
	MaxOutputTokens    *int           `json:"max_output_tokens,omitempty"`
	Temperature        *float64       `json:"temperature,omitempty"`
	TopP               *float64       `json:"top_p,omitempty"`
	Stream             bool           `json:"stream,omitempty"`
	StreamOptions      *StreamOptions `json:"stream_options,omitempty"`
	Tools              []Tool         `json:"tools,omitempty"`
	ToolChoice         any            `json:"tool_choice,omitempty"`
	ParallelToolCalls  *bool          `json:"parallel_tool_calls,omitempty"`
	Reasoning          any            `json:"reasoning,omitempty"`
	Metadata           any            `json:"metadata,omitempty"`
	User               string         `json:"user,omitempty"`
	PreviousResponseID string         `json:"previous_response_id,omitempty"`
	Store              *bool          `json:"store,omitempty"`
	Text               any            `json:"text,omitempty"`
	// 透传的其余字段（如 top_logprobs、truncation 等）不做语义转换
	Extra map[string]any `json:"-"`
}

// ResponsesInputItem input 数组中的单项（message 类型）
type ResponsesInputItem struct {
	Type    string `json:"type,omitempty"` // message / function_call / function_call_output ...
	Role    string `json:"role,omitempty"`
	Content any    `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
	// function_call / function_call_output
	Arguments string `json:"arguments,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Output    string `json:"output,omitempty"`
}

// ResponsesOutputTextContent 输出文本内容块
type ResponsesOutputTextContent struct {
	Type        string `json:"type"` // output_text
	Text        string `json:"text"`
	Annotations any    `json:"annotations,omitempty"`
}

// ResponsesOutputItem 输出项（message 或 function_call）
type ResponsesOutputItem struct {
	Type      string                       `json:"type"` // message / function_call
	Id        string                       `json:"id,omitempty"`
	Role      string                       `json:"role,omitempty"`
	Status    string                       `json:"status,omitempty"`
	Content   []ResponsesOutputTextContent `json:"content,omitempty"`
	Name      string                       `json:"name,omitempty"`
	Arguments string                       `json:"arguments,omitempty"`
	CallID    string                       `json:"call_id,omitempty"`
}

// ResponsesUsage Responses 用量（与 Chat Completions 语义一致）
type ResponsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ResponsesResponse Responses API 响应对象
type ResponsesResponse struct {
	Id        string                `json:"id"`
	Object    string                `json:"object"` // response
	CreatedAt int64                 `json:"created_at"`
	Status    string                `json:"status"` // in_progress / completed / failed
	Model     string                `json:"model"`
	Output    []ResponsesOutputItem `json:"output"`
	Usage     *ResponsesUsage       `json:"usage,omitempty"`
	Error     any                   `json:"error,omitempty"`
}

// ResponsesStreamEvent Responses 流式事件
type ResponsesStreamEvent struct {
	Type           string                      `json:"type"`
	SequenceNumber int                         `json:"sequence_number,omitempty"`
	Response       *ResponsesResponse          `json:"response,omitempty"`
	Item           *ResponsesOutputItem        `json:"item,omitempty"`
	OutputIndex    int                         `json:"output_index,omitempty"`
	ContentIndex   int                         `json:"content_index,omitempty"`
	Delta          string                      `json:"delta,omitempty"`
	Text           string                      `json:"text,omitempty"`
	Part           *ResponsesOutputTextContent `json:"part,omitempty"`
}
