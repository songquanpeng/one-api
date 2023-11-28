package types

type Assistant struct {
	ID           string         `json:"id"`
	Object       string         `json:"object"`
	CreatedAt    int64          `json:"created_at"`
	Name         *string        `json:"name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	Model        string         `json:"model"`
	Instructions *string        `json:"instructions,omitempty"`
	Tools        any            `json:"tools,omitempty"`
	FileIDs      []string       `json:"file_ids,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type AssistantRequest struct {
	Model        string         `json:"model"`
	Name         *string        `json:"name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	Instructions *string        `json:"instructions,omitempty"`
	Tools        any            `json:"tools,omitempty"`
	FileIDs      []string       `json:"file_ids,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// AssistantsList is a list of assistants.
type AssistantsList struct {
	Assistants []Assistant `json:"data"`
	LastID     *string     `json:"last_id"`
	FirstID    *string     `json:"first_id"`
	HasMore    bool        `json:"has_more"`
}

type AssistantDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

type AssistantFile struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	CreatedAt   int64  `json:"created_at"`
	AssistantID string `json:"assistant_id"`
}

type AssistantFileRequest struct {
	FileID string `json:"file_id"`
}

type AssistantFilesList struct {
	AssistantFiles []AssistantFile `json:"data"`
}
