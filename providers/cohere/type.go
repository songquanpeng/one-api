package cohere

import "one-api/types"

type ChatHistory struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type CohereConnector struct {
	ID                string `json:"id"`
	UserAccessToken   string `json:"user_access_token,omitempty"`
	ContinueOnFailure bool   `json:"continue_on_failure,omitempty"`
	Options           any    `json:"options,omitempty"`
}

type CohereRequest struct {
	Message           string                          `json:"message"`
	Model             string                          `json:"model,omitempty"`
	Stream            bool                            `json:"stream,omitempty"`
	Preamble          string                          `json:"preamble,omitempty"`
	ChatHistory       []ChatHistory                   `json:"chat_history,omitempty"`
	ConversationId    string                          `json:"conversation_id,omitempty"`
	PromptTruncation  string                          `json:"prompt_truncation,omitempty"`
	Connectors        []CohereConnector               `json:"connectors,omitempty"`
	Temperature       float64                         `json:"temperature,omitempty"`
	MaxTokens         int                             `json:"max_tokens,omitempty"`
	MaxInputTokens    int                             `json:"max_input_tokens,omitempty"`
	K                 int                             `json:"k,omitempty"`
	P                 float64                         `json:"p,omitempty"`
	Seed              *int                            `json:"seed,omitempty"`
	StopSequences     any                             `json:"stop_sequences,omitempty"`
	FrequencyPenalty  float64                         `json:"frequency_penalty,omitempty"`
	PresencePenalty   float64                         `json:"presence_penalty,omitempty"`
	Tools             []*types.ChatCompletionFunction `json:"tools,omitempty"`
	ToolResults       any                             `json:"tool_results,omitempty"`
	SearchQueriesOnly *bool                           `json:"search_queries_only,omitempty"`
	Documents         []ChatDocument                  `json:"documents,omitempty"`
	CitationQuality   *string                         `json:"citation_quality,omitempty"`
	RawPrompting      *bool                           `json:"raw_prompting,omitempty"`
	ReturnPrompt      *bool                           `json:"return_prompt,omitempty"`
}

type ChatDocument = map[string]string

type APIVersion struct {
	Version string `json:"version"`
}

type Tokens struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	SearchUnits     int `json:"search_units,omitempty"`
	Classifications int `json:"classifications,omitempty"`
}

type Meta struct {
	APIVersion  APIVersion `json:"api_version"`
	BilledUnits Tokens     `json:"billed_units"`
	Tokens      Tokens     `json:"tokens"`
}

type CohereToolCall struct {
	Name       string `json:"name,omitempty"`
	Parameters any    `json:"parameters,omitempty"`
}

type CohereResponse struct {
	Text             string              `json:"text,omitempty"`
	ResponseID       string              `json:"response_id,omitempty"`
	Citations        []*ChatCitation     `json:"citations,omitempty"`
	Documents        []ChatDocument      `json:"documents,omitempty"`
	IsSearchRequired *bool               `json:"is_search_required,omitempty"`
	SearchQueries    []*ChatSearchQuery  `json:"search_queries,omitempty"`
	SearchResults    []*ChatSearchResult `json:"search_results,omitempty"`
	GenerationID     string              `json:"generation_id,omitempty"`
	ChatHistory      []ChatHistory       `json:"chat_history,omitempty"`
	Prompt           *string             `json:"prompt,omitempty"`
	FinishReason     string              `json:"finish_reason,omitempty"`
	ToolCalls        []CohereToolCall    `json:"tool_calls,omitempty"`
	Meta             Meta                `json:"meta,omitempty"`
	CohereError
}

type ChatCitation struct {
	Start       int      `json:"start"`
	End         int      `json:"end"`
	Text        string   `json:"text"`
	DocumentIds []string `json:"document_ids,omitempty"`
}

type ChatSearchQuery struct {
	Text         string `json:"text"`
	GenerationId string `json:"generation_id"`
}

type ChatSearchResult struct {
	SearchQuery       *ChatSearchQuery           `json:"search_query,omitempty" url:"search_query,omitempty"`
	Connector         *ChatSearchResultConnector `json:"connector,omitempty" url:"connector,omitempty"`
	DocumentIds       []string                   `json:"document_ids,omitempty" url:"document_ids,omitempty"`
	ErrorMessage      *string                    `json:"error_message,omitempty" url:"error_message,omitempty"`
	ContinueOnFailure *bool                      `json:"continue_on_failure,omitempty" url:"continue_on_failure,omitempty"`
}

type ChatSearchResultConnector struct {
	Id string `json:"id" url:"id"`
}

type CohereError struct {
	Message string `json:"message,omitempty"`
}

type CohereStreamResponse struct {
	IsFinished   bool             `json:"is_finished"`
	EventType    string           `json:"event_type"`
	GenerationID string           `json:"generation_id,omitempty"`
	Text         string           `json:"text,omitempty"`
	Response     CohereResponse   `json:"response,omitempty"`
	FinishReason string           `json:"finish_reason,omitempty"`
	ToolCalls    []CohereToolCall `json:"tool_calls,omitempty"`
}

type RerankRequest struct {
	Model           *string                       `json:"model,omitempty"`
	Query           string                        `json:"query" url:"query"`
	Documents       []*RerankRequestDocumentsItem `json:"documents,omitempty"`
	TopN            *int                          `json:"top_n,omitempty"`
	RankFields      []string                      `json:"rank_fields,omitempty"`
	ReturnDocuments *bool                         `json:"return_documents,omitempty"`
	MaxChunksPerDoc *int                          `json:"max_chunks_per_doc,omitempty"`
}

type RerankRequestDocumentsItem struct {
	String                         string
	RerankRequestDocumentsItemText *RerankDocumentsItemText
}
type RerankDocumentsItemText struct {
	Text string `json:"text"`
}

type RerankResponse struct {
	Id      *string                      `json:"id,omitempty"`
	Results []*RerankResponseResultsItem `json:"results,omitempty"`
	Meta    *Meta                        `json:"meta,omitempty"`
}

type RerankResponseResultsItem struct {
	Document       *RerankDocumentsItemText `json:"document,omitempty"`
	Index          int                      `json:"index"`
	RelevanceScore float64                  `json:"relevance_score"`
}

type EmbedRequest struct {
	Texts          any      `json:"texts,omitempty"`
	Model          *string  `json:"model,omitempty"`
	InputType      *string  `json:"input_type,omitempty"`
	EmbeddingTypes []string `json:"embedding_types,omitempty"`
	Truncate       *string  `json:"truncate,omitempty"`
}

type EmbedResponse struct {
	ResponseType string `json:"response_type"`
	Embeddings   any    `json:"embeddings"`
}

type EmbedFloatsResponse struct {
	Id         string      `json:"id"`
	Embeddings [][]float64 `json:"embeddings,omitempty"`
	Texts      []string    `json:"texts,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
}

type EmbedByTypeResponse struct {
	Id         string                         `json:"id"`
	Embeddings *EmbedByTypeResponseEmbeddings `json:"embeddings,omitempty"`
	Texts      []string                       `json:"texts,omitempty"`
	Meta       *Meta                          `json:"meta,omitempty"`
}

type EmbedByTypeResponseEmbeddings struct {
	Float   [][]float64 `json:"float,omitempty"`
	Int8    [][]int     `json:"int8,omitempty"`
	Uint8   [][]int     `json:"uint8,omitempty"`
	Binary  [][]int     `json:"binary,omitempty"`
	Ubinary [][]int     `json:"ubinary,omitempty"`
}

type ModelListResponse struct {
	Models []ModelDetails `json:"models"`
}

type ModelDetails struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}
