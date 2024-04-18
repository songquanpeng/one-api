package stabilityAI

import "strings"

type StabilityAIError struct {
	Name    string   `json:"name,omitempty"`
	Errors  []string `json:"errors,omitempty"`
	Success bool     `json:"success,omitempty"`
	Message string   `json:"message,omitempty"`
}

func (e StabilityAIError) String() string {
	return strings.Join(e.Errors, ", ")
}

type generateResponse struct {
	Image        string `json:"image"`
	FinishReason string `json:"finish_reason,omitempty"`
	Seed         int    `json:"seed,omitempty"`
}
