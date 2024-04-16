package types

import "mime/multipart"

type SpeechAudioRequest struct {
	Model          string  `json:"model" binding:"required"`
	Input          string  `json:"input" binding:"required"`
	Voice          string  `json:"voice" binding:"required"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

type AudioRequest struct {
	File           *multipart.FileHeader `form:"file" binding:"required"`
	Model          string                `form:"model" binding:"required"`
	Language       string                `form:"language"`
	Prompt         string                `form:"prompt"`
	ResponseFormat string                `form:"response_format"`
	Temperature    float32               `form:"temperature"`
}

type AudioResponse struct {
	Task     string           `json:"task,omitempty"`
	Language string           `json:"language,omitempty"`
	Duration float64          `json:"duration,omitempty"`
	Segments any              `json:"segments,omitempty"`
	Text     string           `json:"text"`
	Words    []AudioWordsList `json:"words,omitempty"`
}

type AudioWordsList struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type AudioResponseWrapper struct {
	Headers map[string]string
	Body    []byte
}
