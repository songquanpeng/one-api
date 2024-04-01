package model

type Tool struct {
	Id       string   `json:"id,omitempty"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name"`
	Parameters  any    `json:"parameters,omitempty"` // request
	Arguments   any    `json:"arguments,omitempty"`  // response
}
