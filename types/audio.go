package types

import "mime/multipart"

type SpeechAudioRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

type AudioRequest struct {
	File           *multipart.FileHeader `form:"file"`
	Model          string                `form:"model"`
	Language       string                `form:"language"`
	Prompt         string                `form:"prompt"`
	ResponseFormat string                `form:"response_format"`
	Temperature    float32               `form:"temperature"`
}

type AudioResponse struct {
	Task     string  `json:"task,omitempty"`
	Language string  `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Segments any     `json:"segments,omitempty"`
	Text     string  `json:"text"`
}
