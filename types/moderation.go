package types

type ModerationRequest struct {
	Input string `json:"input,omitempty"`
	Model string `json:"model,omitempty"`
}

type ModerationResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Results any    `json:"results"`
}
