package vertexai

type EmbeddingTaskType string

const (
	EmbeddingTaskTypeNone               EmbeddingTaskType = ""
	EmbeddingTaskTypeRetrievalQuery     EmbeddingTaskType = "RETRIEVAL_QUERY"
	EmbeddingTaskTypeRetrievalDocument  EmbeddingTaskType = "RETRIEVAL_DOCUMENT"
	EmbeddingTaskTypeSemanticSimilarity EmbeddingTaskType = "SEMANTIC_SIMILARITY"
	EmbeddingTaskTypeClassification     EmbeddingTaskType = "CLASSIFICATION"
	EmbeddingTaskTypeClustering         EmbeddingTaskType = "CLUSTERING"
	EmbeddingTaskTypeQuestionAnswering  EmbeddingTaskType = "QUESTION_ANSWERING"
	EmbeddingTaskTypeFactVerification   EmbeddingTaskType = "FACT_VERIFICATION"
	EmbeddingTaskTypeCodeRetrievalQuery EmbeddingTaskType = "CODE_RETRIEVAL_QUERY"
)

type EmbeddingRequest struct {
	Instances  []EmbeddingInstance `json:"instances"`
	Parameters EmbeddingParams     `json:"parameters"`
}

type EmbeddingInstance struct {
	Content  string            `json:"content"`
	TaskType EmbeddingTaskType `json:"task_type,omitempty"`
	Title    string            `json:"title,omitempty"`
}

type EmbeddingParams struct {
	AutoTruncate         bool `json:"autoTruncate,omitempty"`
	OutputDimensionality int  `json:"outputDimensionality,omitempty"`
	// Texts                []string `json:"texts,omitempty"`
}

type EmbeddingResponse struct {
	Predictions []struct {
		Embeddings EmbeddingData `json:"embeddings"`
	} `json:"predictions"`
}

type EmbeddingData struct {
	Statistics struct {
		Truncated  bool `json:"truncated"`
		TokenCount int  `json:"token_count"`
	} `json:"statistics"`
	Values []float64 `json:"values"`
}
