package anthropic

import (
	"encoding/json"
	"fmt"
)

// https://docs.anthropic.com/claude/reference/messages_post

type Metadata struct {
	UserId string `json:"user_id"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Content struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
	// tool_calls
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Input     any    `json:"input,omitempty"`
	Content   string `json:"content,omitempty"`
	ToolUseId string `json:"tool_use_id,omitempty"`
}

// MessageContent can handle both string and array formats for the content field
type MessageContent struct {
	value interface{}
}

// UnmarshalJSON implements json.Unmarshaler to handle both string and array formats
func (m *MessageContent) UnmarshalJSON(data []byte) error {
	// Skip empty data or null
	if len(data) == 0 || string(data) == "null" {
		m.value = ""
		return nil
	}

	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		m.value = str
		return nil
	}

	// If that fails, try to unmarshal as array of Content
	var arr []Content
	if err := json.Unmarshal(data, &arr); err == nil {
		m.value = arr
		return nil
	}

	// For routing purposes, store raw JSON if we can't parse it
	// This ensures we don't lose any data during forwarding
	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		m.value = raw
		return nil
	}

	return fmt.Errorf("message content must be either a string or an array of content blocks")
}

// MarshalJSON implements json.Marshaler
func (m MessageContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.value)
}

// ToContentArray converts the message content to a []Content array
func (m MessageContent) ToContentArray() []Content {
	if m.value == nil {
		return []Content{}
	}

	switch v := m.value.(type) {
	case string:
		// Convert string to a single text content block
		return []Content{{
			Type: "text",
			Text: v,
		}}
	case []Content:
		return v
	case json.RawMessage:
		// Try to parse raw JSON as Content array
		var arr []Content
		if err := json.Unmarshal(v, &arr); err == nil {
			return arr
		}
		// If that fails, return empty array to avoid breaking the routing
		return []Content{}
	default:
		return []Content{}
	}
}

type Message struct {
	Role    string         `json:"role"`
	Content MessageContent `json:"content"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string `json:"type"`
	Properties any    `json:"properties,omitempty"`
	Required   any    `json:"required,omitempty"`
}

// SystemPrompt can handle both string and array formats for the system field
type SystemPrompt struct {
	value interface{}
}

// UnmarshalJSON implements json.Unmarshaler to handle both string and array formats
func (s *SystemPrompt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.value = str
		return nil
	}

	// If that fails, try to unmarshal as array
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		s.value = arr
		return nil
	}

	return fmt.Errorf("system field must be either a string or an array")
}

// MarshalJSON implements json.Marshaler
func (s SystemPrompt) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

// String returns the system prompt as a string
func (s SystemPrompt) String() string {
	if s.value == nil {
		return ""
	}

	switch v := s.value.(type) {
	case string:
		return v
	case []interface{}:
		// Convert array to string by concatenating text content
		var result string
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, exists := itemMap["text"]; exists {
					if textStr, ok := text.(string); ok {
						result += textStr + " "
					}
				}
			} else if str, ok := item.(string); ok {
				result += str + " "
			}
		}
		return result
	default:
		return fmt.Sprintf("%v", v)
	}
}

// IsEmpty returns true if the system prompt is empty
func (s SystemPrompt) IsEmpty() bool {
	if s.value == nil {
		return true
	}

	switch v := s.value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	default:
		return false
	}
}

type Request struct {
	Model         string       `json:"model"`
	Messages      []Message    `json:"messages"`
	System        SystemPrompt `json:"system,omitempty"`
	MaxTokens     int          `json:"max_tokens,omitempty"`
	StopSequences []string     `json:"stop_sequences,omitempty"`
	Stream        bool         `json:"stream,omitempty"`
	Temperature   *float64     `json:"temperature,omitempty"`
	TopP          *float64     `json:"top_p,omitempty"`
	TopK          int          `json:"top_k,omitempty"`
	Tools         []Tool       `json:"tools,omitempty"`
	ToolChoice    any          `json:"tool_choice,omitempty"`
	//Metadata    `json:"metadata,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Response struct {
	Id           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      MessageContent `json:"content"`
	Model        string         `json:"model"`
	StopReason   *string        `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        Usage          `json:"usage"`
	Error        Error          `json:"error"`
}

type Delta struct {
	Type         string  `json:"type"`
	Text         string  `json:"text"`
	PartialJson  string  `json:"partial_json,omitempty"`
	StopReason   *string `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
}

type StreamResponse struct {
	Type         string    `json:"type"`
	Message      *Response `json:"message"`
	Index        int       `json:"index"`
	ContentBlock *Content  `json:"content_block"`
	Delta        *Delta    `json:"delta"`
	Usage        *Usage    `json:"usage"`
}
