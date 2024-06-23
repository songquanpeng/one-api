package model

import "math"

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (u *Usage) Quota(completionRatio, finalRatio float64) int64 {
	quota := int64(math.Ceil((float64(u.PromptTokens) + float64(u.CompletionTokens)*completionRatio) * finalRatio))
	if finalRatio != 0 && quota <= 0 {
		quota = 1
	}
	return quota
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
}

type ErrorWithStatusCode struct {
	Error
	StatusCode int `json:"status_code"`
}
