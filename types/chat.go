package types

const (
	ContentTypeText     = "text"
	ContentTypeImageURL = "image_url"
)

const (
	FinishReasonStop          = "stop"
	FinishReasonLength        = "length"
	FinishReasonFunctionCall  = "function_call"
	FinishReasonToolCalls     = "tool_calls"
	FinishReasonContentFilter = "content_filter"
	FinishReasonNull          = "null"
)

const (
	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleFunction  = "function"
	ChatMessageRoleTool      = "tool"
)

type ChatCompletionToolCallsFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments"`
}

type ChatCompletionToolCalls struct {
	Id       string                           `json:"id"`
	Type     string                           `json:"type"`
	Function *ChatCompletionToolCallsFunction `json:"function"`
	Index    int                              `json:"index"`
}

type ChatCompletionMessage struct {
	Role         string                           `json:"role"`
	Content      any                              `json:"content,omitempty"`
	Name         *string                          `json:"name,omitempty"`
	FunctionCall *ChatCompletionToolCallsFunction `json:"function_call,omitempty"`
	ToolCalls    []*ChatCompletionToolCalls       `json:"tool_calls,omitempty"`
	ToolCallID   string                           `json:"tool_call_id,omitempty"`
}

func (m ChatCompletionMessage) StringContent() string {
	content, ok := m.Content.(string)
	if ok {
		return content
	}
	contentList, ok := m.Content.([]any)
	if ok {
		var contentStr string
		for _, contentItem := range contentList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}

			if subStr, ok := contentMap["text"].(string); ok && subStr != "" {
				contentStr += subStr
			}

		}
		return contentStr
	}
	return ""
}

func (m ChatCompletionMessage) ParseContent() []ChatMessagePart {
	var contentList []ChatMessagePart
	content, ok := m.Content.(string)
	if ok {
		contentList = append(contentList, ChatMessagePart{
			Type: ContentTypeText,
			Text: content,
		})
		return contentList
	}
	anyList, ok := m.Content.([]any)
	if ok {
		for _, contentItem := range anyList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}

			if subStr, ok := contentMap["text"].(string); ok && subStr != "" {
				contentList = append(contentList, ChatMessagePart{
					Type: ContentTypeText,
					Text: subStr,
				})
			} else if subObj, ok := contentMap["image_url"].(map[string]any); ok {
				contentList = append(contentList, ChatMessagePart{
					Type: ContentTypeImageURL,
					ImageURL: &ChatMessageImageURL{
						URL: subObj["url"].(string),
					},
				})
			} else if subObj, ok := contentMap["image"].(string); ok {
				contentList = append(contentList, ChatMessagePart{
					Type: ContentTypeImageURL,
					ImageURL: &ChatMessageImageURL{
						URL: subObj,
					},
				})
			}
		}
		return contentList
	}
	return nil
}

// 将FunctionCall转换为ToolCalls
func (m *ChatCompletionMessage) FuncToToolCalls() {
	if m.ToolCalls != nil {
		return
	}
	if m.FunctionCall != nil {
		m.ToolCalls = []*ChatCompletionToolCalls{
			{
				Type:     ChatMessageRoleFunction,
				Function: m.FunctionCall,
			},
		}
		m.FunctionCall = nil
	}
}

// 将ToolCalls转换为FunctionCall
func (m *ChatCompletionMessage) ToolToFuncCalls() {

	if m.FunctionCall != nil {
		return
	}
	if m.ToolCalls != nil {
		m.FunctionCall = &ChatCompletionToolCallsFunction{
			Name:      m.ToolCalls[0].Function.Name,
			Arguments: m.ToolCalls[0].Function.Arguments,
		}
		m.ToolCalls = nil
	}
}

type ChatMessageImageURL struct {
	URL    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type ChatMessagePart struct {
	Type     string               `json:"type,omitempty"`
	Text     string               `json:"text,omitempty"`
	ImageURL *ChatMessageImageURL `json:"image_url,omitempty"`
}

type ChatCompletionResponseFormat struct {
	Type string `json:"type,omitempty"`
}

type ChatCompletionRequest struct {
	Model            string                        `json:"model" binding:"required"`
	Messages         []ChatCompletionMessage       `json:"messages" binding:"required"`
	MaxTokens        int                           `json:"max_tokens,omitempty"`
	Temperature      float64                       `json:"temperature,omitempty"`
	TopP             float64                       `json:"top_p,omitempty"`
	N                int                           `json:"n,omitempty"`
	Stream           bool                          `json:"stream,omitempty"`
	StreamOptions    *StreamOptions                `json:"stream_options,omitempty"`
	Stop             []string                      `json:"stop,omitempty"`
	PresencePenalty  float64                       `json:"presence_penalty,omitempty"`
	ResponseFormat   *ChatCompletionResponseFormat `json:"response_format,omitempty"`
	Seed             *int                          `json:"seed,omitempty"`
	FrequencyPenalty float64                       `json:"frequency_penalty,omitempty"`
	LogitBias        any                           `json:"logit_bias,omitempty"`
	LogProbs         *bool                         `json:"logprobs,omitempty"`
	TopLogProbs      int                           `json:"top_logprobs,omitempty"`
	User             string                        `json:"user,omitempty"`
	Functions        []*ChatCompletionFunction     `json:"functions,omitempty"`
	FunctionCall     any                           `json:"function_call,omitempty"`
	Tools            []*ChatCompletionTool         `json:"tools,omitempty"`
	ToolChoice       any                           `json:"tool_choice,omitempty"`
}

func (r ChatCompletionRequest) GetFunctionCate() string {
	if r.Tools != nil {
		return "tool"
	} else if r.Functions != nil {
		return "function"
	}
	return ""
}

func (r *ChatCompletionRequest) GetFunctions() []*ChatCompletionFunction {
	if r.Tools == nil && r.Functions == nil {
		return nil
	}

	if r.Tools != nil {
		var functions []*ChatCompletionFunction
		for _, tool := range r.Tools {
			functions = append(functions, &tool.Function)
		}
		return functions
	}

	return r.Functions
}

func (r *ChatCompletionRequest) ClearEmptyMessages() {
	var messages []ChatCompletionMessage
	for _, message := range r.Messages {
		if message.StringContent() != "" || message.ToolCalls != nil || message.FunctionCall != nil {
			messages = append(messages, message)
		}
	}
	r.Messages = messages
}

type ChatCompletionFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type ChatCompletionTool struct {
	Type     string                 `json:"type"`
	Function ChatCompletionFunction `json:"function"`
}

type ChatCompletionChoice struct {
	Index                int                   `json:"index"`
	Message              ChatCompletionMessage `json:"message"`
	LogProbs             any                   `json:"logprobs,omitempty"`
	FinishReason         any                   `json:"finish_reason,omitempty"`
	ContentFilterResults any                   `json:"content_filter_results,omitempty"`
	FinishDetails        any                   `json:"finish_details,omitempty"`
}

func (c *ChatCompletionChoice) CheckChoice(request *ChatCompletionRequest) {
	if request.Functions != nil && c.Message.ToolCalls != nil {
		c.Message.ToolToFuncCalls()
		c.FinishReason = FinishReasonFunctionCall
	}
}

type ChatCompletionResponse struct {
	ID                  string                 `json:"id"`
	Object              string                 `json:"object"`
	Created             any                    `json:"created"`
	Model               string                 `json:"model"`
	Choices             []ChatCompletionChoice `json:"choices"`
	Usage               *Usage                 `json:"usage,omitempty"`
	SystemFingerprint   string                 `json:"system_fingerprint,omitempty"`
	PromptFilterResults any                    `json:"prompt_filter_results,omitempty"`
}

func (cc *ChatCompletionResponse) GetContent() string {
	var content string
	for _, choice := range cc.Choices {
		content += choice.Message.StringContent()
	}
	return content
}

func (c ChatCompletionStreamChoice) ConvertOpenaiStream() []ChatCompletionStreamChoice {
	var choices []ChatCompletionStreamChoice
	var stopFinish string
	if c.Delta.FunctionCall != nil {
		stopFinish = FinishReasonFunctionCall
		choices = c.Delta.FunctionCall.Split(&c, stopFinish, 0)
	} else {
		stopFinish = FinishReasonToolCalls
		for index, tool := range c.Delta.ToolCalls {
			choices = append(choices, tool.Function.Split(&c, stopFinish, index)...)
		}
	}

	choices = append(choices, ChatCompletionStreamChoice{
		Index:        c.Index,
		Delta:        ChatCompletionStreamChoiceDelta{},
		FinishReason: stopFinish,
	})

	return choices
}

func (f *ChatCompletionToolCallsFunction) Split(c *ChatCompletionStreamChoice, stopFinish string, index int) []ChatCompletionStreamChoice {
	var functions []*ChatCompletionToolCallsFunction
	var choices []ChatCompletionStreamChoice
	functions = append(functions, &ChatCompletionToolCallsFunction{
		Name:      f.Name,
		Arguments: "",
	})

	if f.Arguments == "" || f.Arguments == "{}" {
		functions = append(functions, &ChatCompletionToolCallsFunction{
			Arguments: "{}",
		})
	} else {
		functions = append(functions, &ChatCompletionToolCallsFunction{
			Arguments: f.Arguments,
		})
	}

	for fIndex, function := range functions {
		choice := ChatCompletionStreamChoice{
			Index: c.Index,
			Delta: ChatCompletionStreamChoiceDelta{
				Role: c.Delta.Role,
			},
		}
		if stopFinish == FinishReasonFunctionCall {
			choice.Delta.FunctionCall = function
		} else {
			toolCalls := &ChatCompletionToolCalls{
				// Id:       c.Delta.ToolCalls[0].Id,
				Index:    index,
				Type:     ChatMessageRoleFunction,
				Function: function,
			}

			if fIndex == 0 {
				toolCalls.Id = c.Delta.ToolCalls[0].Id
			}
			choice.Delta.ToolCalls = []*ChatCompletionToolCalls{toolCalls}
		}

		choices = append(choices, choice)
	}

	return choices
}

type ChatCompletionStreamChoiceDelta struct {
	Content      string                           `json:"content,omitempty"`
	Role         string                           `json:"role,omitempty"`
	FunctionCall *ChatCompletionToolCallsFunction `json:"function_call,omitempty"`
	ToolCalls    []*ChatCompletionToolCalls       `json:"tool_calls,omitempty"`
}

func (m *ChatCompletionStreamChoiceDelta) ToolToFuncCalls() {

	if m.FunctionCall != nil {
		return
	}
	if m.ToolCalls != nil {
		m.FunctionCall = &ChatCompletionToolCallsFunction{
			Name:      m.ToolCalls[0].Function.Name,
			Arguments: m.ToolCalls[0].Function.Arguments,
		}
		m.ToolCalls = nil
	}
}

type ChatCompletionStreamChoice struct {
	Index                int                             `json:"index"`
	Delta                ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason         any                             `json:"finish_reason"`
	ContentFilterResults any                             `json:"content_filter_results,omitempty"`
	Usage                *Usage                          `json:"usage,omitempty"`
}

func (c *ChatCompletionStreamChoice) CheckChoice(request *ChatCompletionRequest) {
	if request.Functions != nil && c.Delta.ToolCalls != nil {
		c.Delta.ToolToFuncCalls()
		c.FinishReason = FinishReasonToolCalls
	}
}

type ChatCompletionStreamResponse struct {
	ID                string                       `json:"id"`
	Object            string                       `json:"object"`
	Created           any                          `json:"created"`
	Model             string                       `json:"model"`
	Choices           []ChatCompletionStreamChoice `json:"choices"`
	PromptAnnotations any                          `json:"prompt_annotations,omitempty"`
	Usage             *Usage                       `json:"usage,omitempty"`
}

func (c *ChatCompletionStreamResponse) GetResponseText() (responseText string) {
	for _, choice := range c.Choices {
		responseText += choice.Delta.Content
	}

	return
}
