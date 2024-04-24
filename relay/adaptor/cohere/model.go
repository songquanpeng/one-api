package cohere

type Request struct {
	Message          string        `json:"message" required:"true"`
	Model            string        `json:"model,omitempty"`  // 默认值为"command-r"
	Stream           bool          `json:"stream,omitempty"` // 默认值为false
	Preamble         string        `json:"preamble,omitempty"`
	ChatHistory      []ChatMessage `json:"chat_history,omitempty"`
	ConversationID   string        `json:"conversation_id,omitempty"`
	PromptTruncation string        `json:"prompt_truncation,omitempty"` // 默认值为"AUTO"
	Connectors       []Connector   `json:"connectors,omitempty"`
	Documents        []Document    `json:"documents,omitempty"`
	Temperature      float64       `json:"temperature,omitempty"` // 默认值为0.3
	MaxTokens        int           `json:"max_tokens,omitempty"`
	MaxInputTokens   int           `json:"max_input_tokens,omitempty"`
	K                int           `json:"k,omitempty"` // 默认值为0
	P                float64       `json:"p,omitempty"` // 默认值为0.75
	Seed             int           `json:"seed,omitempty"`
	StopSequences    []string      `json:"stop_sequences,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"` // 默认值为0.0
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`  // 默认值为0.0
	Tools            []Tool        `json:"tools,omitempty"`
	ToolResults      []ToolResult  `json:"tool_results,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role" required:"true"`
	Message string `json:"message" required:"true"`
}

type Tool struct {
	Name                 string                   `json:"name" required:"true"`
	Description          string                   `json:"description" required:"true"`
	ParameterDefinitions map[string]ParameterSpec `json:"parameter_definitions"`
}

type ParameterSpec struct {
	Description string `json:"description"`
	Type        string `json:"type" required:"true"`
	Required    bool   `json:"required"`
}

type ToolResult struct {
	Call    ToolCall                 `json:"call"`
	Outputs []map[string]interface{} `json:"outputs"`
}

type ToolCall struct {
	Name       string                 `json:"name" required:"true"`
	Parameters map[string]interface{} `json:"parameters" required:"true"`
}

type StreamResponse struct {
	IsFinished    bool            `json:"is_finished"`
	EventType     string          `json:"event_type"`
	GenerationID  string          `json:"generation_id,omitempty"`
	SearchQueries []*SearchQuery  `json:"search_queries,omitempty"`
	SearchResults []*SearchResult `json:"search_results,omitempty"`
	Documents     []*Document     `json:"documents,omitempty"`
	Text          string          `json:"text,omitempty"`
	Citations     []*Citation     `json:"citations,omitempty"`
	Response      *Response       `json:"response,omitempty"`
	FinishReason  string          `json:"finish_reason,omitempty"`
}

type SearchQuery struct {
	Text         string `json:"text"`
	GenerationID string `json:"generation_id"`
}

type SearchResult struct {
	SearchQuery *SearchQuery `json:"search_query"`
	DocumentIDs []string     `json:"document_ids"`
	Connector   *Connector   `json:"connector"`
}

type Connector struct {
	ID string `json:"id"`
}

type Document struct {
	ID        string `json:"id"`
	Snippet   string `json:"snippet"`
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
	URL       string `json:"url"`
}

type Citation struct {
	Start       int      `json:"start"`
	End         int      `json:"end"`
	Text        string   `json:"text"`
	DocumentIDs []string `json:"document_ids"`
}

type Response struct {
	ResponseID    string          `json:"response_id"`
	Text          string          `json:"text"`
	GenerationID  string          `json:"generation_id"`
	ChatHistory   []*Message      `json:"chat_history"`
	FinishReason  *string         `json:"finish_reason"`
	Meta          Meta            `json:"meta"`
	Citations     []*Citation     `json:"citations"`
	Documents     []*Document     `json:"documents"`
	SearchResults []*SearchResult `json:"search_results"`
	SearchQueries []*SearchQuery  `json:"search_queries"`
	Message       string          `json:"message"`
}

type Message struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type Version struct {
	Version string `json:"version"`
}

type Units struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type ChatEntry struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type Meta struct {
	APIVersion  APIVersion  `json:"api_version"`
	BilledUnits BilledUnits `json:"billed_units"`
	Tokens      Usage       `json:"tokens"`
}

type APIVersion struct {
	Version string `json:"version"`
}

type BilledUnits struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
