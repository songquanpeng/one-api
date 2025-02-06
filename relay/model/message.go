package model

import (
	"context"

	"github.com/songquanpeng/one-api/common/logger"
)

type Message struct {
	Role string `json:"role,omitempty"`
	// Content is a string or a list of objects
	Content    any           `json:"content,omitempty"`
	Name       *string       `json:"name,omitempty"`
	ToolCalls  []Tool        `json:"tool_calls,omitempty"`
	ToolCallId string        `json:"tool_call_id,omitempty"`
	Audio      *messageAudio `json:"audio,omitempty"`
	// -------------------------------------
	// Deepseek 专有的一些字段
	// https://api-docs.deepseek.com/api/create-chat-completion
	// -------------------------------------
	// Prefix forces the model to begin its answer with the supplied prefix in the assistant message.
	// To enable this feature, set base_url to "https://api.deepseek.com/beta".
	Prefix *bool `json:"prefix,omitempty"` // ReasoningContent is Used for the deepseek-reasoner model in the Chat
	// Prefix Completion feature as the input for the CoT in the last assistant message.
	// When using this feature, the prefix parameter must be set to true.
	ReasoningContent *string `json:"reasoning_content,omitempty"`
}

type messageAudio struct {
	Id         string `json:"id"`
	Data       string `json:"data,omitempty"`
	ExpiredAt  int    `json:"expired_at,omitempty"`
	Transcript string `json:"transcript,omitempty"`
}

func (m Message) IsStringContent() bool {
	_, ok := m.Content.(string)
	return ok
}

func (m Message) StringContent() string {
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

			if contentMap["type"] == ContentTypeText {
				if subStr, ok := contentMap["text"].(string); ok {
					contentStr += subStr
				}
			}
		}
		return contentStr
	}

	return ""
}

func (m Message) ParseContent() []MessageContent {
	var contentList []MessageContent
	content, ok := m.Content.(string)
	if ok {
		contentList = append(contentList, MessageContent{
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
			switch contentMap["type"] {
			case ContentTypeText:
				if subStr, ok := contentMap["text"].(string); ok {
					contentList = append(contentList, MessageContent{
						Type: ContentTypeText,
						Text: subStr,
					})
				}
			case ContentTypeImageURL:
				if subObj, ok := contentMap["image_url"].(map[string]any); ok {
					contentList = append(contentList, MessageContent{
						Type: ContentTypeImageURL,
						ImageURL: &ImageURL{
							Url: subObj["url"].(string),
						},
					})
				}
			case ContentTypeInputAudio:
				if subObj, ok := contentMap["input_audio"].(map[string]any); ok {
					contentList = append(contentList, MessageContent{
						Type: ContentTypeInputAudio,
						InputAudio: &InputAudio{
							Data:   subObj["data"].(string),
							Format: subObj["format"].(string),
						},
					})
				}
			default:
				logger.Warnf(context.TODO(), "unknown content type: %s", contentMap["type"])
			}
		}

		return contentList
	}
	return nil
}

type ImageURL struct {
	Url    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type MessageContent struct {
	// Type should be one of the following: text/input_audio
	Type       string      `json:"type,omitempty"`
	Text       string      `json:"text"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
}

type InputAudio struct {
	// Data is the base64 encoded audio data
	Data string `json:"data" binding:"required"`
	// Format is the audio format, should be one of the
	// following: mp3/mp4/mpeg/mpga/m4a/wav/webm/pcm16.
	// When stream=true, format should be pcm16
	Format string `json:"format"`
}
