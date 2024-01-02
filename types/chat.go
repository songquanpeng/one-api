package types

const (
	ContentTypeText     = "text"
	ContentTypeImageURL = "image_url"
)

type ChatCompletionToolCallsFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ChatCompletionToolCalls struct {
	Id       string                          `json:"id"`
	Type     string                          `json:"type"`
	Function ChatCompletionToolCallsFunction `json:"function"`
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
	Stop             []string                      `json:"stop,omitempty"`
	PresencePenalty  float64                       `json:"presence_penalty,omitempty"`
	ResponseFormat   *ChatCompletionResponseFormat `json:"response_format,omitempty"`
	Seed             *int                          `json:"seed,omitempty"`
	FrequencyPenalty float64                       `json:"frequency_penalty,omitempty"`
	LogitBias        any                           `json:"logit_bias,omitempty"`
	User             string                        `json:"user,omitempty"`
	Functions        []*ChatCompletionFunction     `json:"functions,omitempty"`
	FunctionCall     any                           `json:"function_call,omitempty"`
	Tools            []*ChatCompletionTool         `json:"tools,omitempty"`
	ToolChoice       any                           `json:"tool_choice,omitempty"`
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
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason any                   `json:"finish_reason,omitempty"`
}

type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             *Usage                 `json:"usage,omitempty"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
}

type ChatCompletionStreamChoiceDelta struct {
	Content      string `json:"content,omitempty"`
	Role         string `json:"role,omitempty"`
	FunctionCall any    `json:"function_call,omitempty"`
	ToolCalls    any    `json:"tool_calls,omitempty"`
}

type ChatCompletionStreamChoice struct {
	Index                int                             `json:"index"`
	Delta                ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason         any                             `json:"finish_reason"`
	ContentFilterResults any                             `json:"content_filter_results,omitempty"`
}

type ChatCompletionStreamResponse struct {
	ID                string                       `json:"id"`
	Object            string                       `json:"object"`
	Created           int64                        `json:"created"`
	Model             string                       `json:"model"`
	Choices           []ChatCompletionStreamChoice `json:"choices"`
	PromptAnnotations any                          `json:"prompt_annotations,omitempty"`
}
