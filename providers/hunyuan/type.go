package hunyuan

type HunyuanError struct {
	Message string `json:"Message,omitempty" name:"Message"`
	Code    string `json:"Code,omitempty" name:"Code"`
}

type ChatCompletionsRequest struct {
	Model             string     `json:"Model,omitempty" name:"Model"`
	Messages          []*Message `json:"Messages,omitempty" name:"Messages"`
	Stream            bool       `json:"Stream,omitempty" name:"Stream"`
	StreamModeration  *bool      `json:"StreamModeration,omitempty" name:"StreamModeration"`
	TopP              *float64   `json:"TopP,omitempty" name:"TopP"`
	Temperature       *float64   `json:"Temperature,omitempty" name:"Temperature"`
	EnableEnhancement *bool      `json:"EnableEnhancement,omitempty" name:"EnableEnhancement"`
}

type Message struct {
	Role    string `json:"Role,omitempty" name:"Role"`
	Content string `json:"Content,omitempty" name:"Content"`
}

type ChatCompletionsResponse struct {
	Response *ChatCompletionsResponseParams `json:"Response"`
}

type ChatCompletionsResponseParams struct {
	// Unix 时间戳，单位为秒。
	Created *int64 `json:"Created,omitempty" name:"Created"`

	// Token 统计信息。
	// 按照总 Token 数量计费。
	Usage *HunyuanUsage `json:"Usage,omitempty" name:"Usage"`

	// 免责声明。
	Note *string `json:"Note,omitempty" name:"Note"`

	// 本轮对话的 ID。
	Id *string `json:"Id,omitempty" name:"Id"`

	// 回复内容。
	Choices []*Choice `json:"Choices,omitempty" name:"Choices"`

	// 唯一请求 ID，由服务端生成，每次请求都会返回（若请求因其他原因未能抵达服务端，则该次请求不会获得 RequestId）。定位问题时需要提供该次请求的 RequestId。本接口为流式响应接口，当请求成功时，RequestId 会被放在 HTTP 响应的 Header "X-TC-RequestId" 中。
	RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	HunyuanResponseError
}

type Choice struct {
	// 结束标志位，可能为 stop 或 sensitive。
	// stop 表示输出正常结束，sensitive 只在开启流式输出审核时会出现，表示安全审核未通过。
	FinishReason *string `json:"FinishReason,omitempty" name:"FinishReason"`

	// 增量返回值，流式调用时使用该字段。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Delta *Delta `json:"Delta,omitempty" name:"Delta"`

	// 返回值，非流式调用时使用该字段。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Message *Message `json:"Message,omitempty" name:"Message"`
}

type Delta struct {
	// 角色名称。
	Role string `json:"Role,omitempty" name:"Role"`

	// 内容详情。
	Content string `json:"Content,omitempty" name:"Content"`
}

type HunyuanUsage struct {
	// 输入 Token 数量。
	PromptTokens int `json:"PromptTokens,omitempty" name:"PromptTokens"`

	// 输出 Token 数量。
	CompletionTokens int `json:"CompletionTokens,omitempty" name:"CompletionTokens"`

	// 总 Token 数量。
	TotalTokens int `json:"TotalTokens,omitempty" name:"TotalTokens"`
}

type HunyuanResponseError struct {
	// 错误信息。
	// 如果流式返回中服务处理异常，返回该错误信息。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Error *HunyuanError `json:"Error,omitempty" name:"Error"`
}
