package model

type RerankRequest struct {
	Model           string   `json:"model"`
	Documents       []string `json:"documents"`
	Query           string   `json:"query"`
	TopN            *int     `json:"top_n,omitempty"`
	MaxChunksPerDoc *int     `json:"max_chunks_per_doc,omitempty"`
	ReturnDocuments *bool    `json:"return_documents,omitempty"`
}
