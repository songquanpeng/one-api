package model

type ImageRequest struct {
	Model          string `json:"model" form:"model"`
	Prompt         string `json:"prompt" form:"prompt" binding:"required"`
	N              int    `json:"n,omitempty" form:"n"`
	Size           string `json:"size,omitempty" form:"size"`
	Quality        string `json:"quality,omitempty" form:"quality"`
	ResponseFormat string `json:"response_format,omitempty" form:"response_format"`
	Style          string `json:"style,omitempty" form:"style"`
	User           string `json:"user,omitempty" form:"user"`
}
