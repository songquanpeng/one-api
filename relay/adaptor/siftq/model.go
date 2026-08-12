package siftq

type URLValue struct {
	URL string `json:"url"`
}

type ContentItem struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *URLValue `json:"image_url,omitempty"`
	VideoURL *URLValue `json:"video_url,omitempty"`
	AudioURL *URLValue `json:"audio_url,omitempty"`
	Role     string    `json:"role,omitempty"`
}

type CreateRequest struct {
	Model       string        `json:"model"`
	Content     []ContentItem `json:"content"`
	Resolution  *string       `json:"resolution,omitempty"`
	Duration    *int          `json:"duration"`
	Ratio       *string       `json:"ratio,omitempty"`
	CallbackURL string        `json:"callback_url,omitempty"`
}

type CreateResponse struct {
	TaskID string `json:"task_id"`
}

type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TaskContent struct {
	URL    string `json:"url,omitempty"`
	Prompt string `json:"prompt,omitempty"`
}

type TaskUsage struct {
	TotalSeconds     int `json:"total_seconds,omitempty"`
	InputSeconds     int `json:"input_seconds,omitempty"`
	OutputSeconds    int `json:"output_seconds,omitempty"`
	InputImageCount  int `json:"input_image_count,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
}

type VideoTask struct {
	ID         string       `json:"id"`
	Model      string       `json:"model,omitempty"`
	Status     string       `json:"status"`
	Error      *TaskError   `json:"error,omitempty"`
	CreatedAt  int64        `json:"created_at,omitempty"`
	UpdatedAt  int64        `json:"updated_at,omitempty"`
	Content    *TaskContent `json:"content,omitempty"`
	Resolution string       `json:"resolution,omitempty"`
	Duration   int          `json:"duration,omitempty"`
	Ratio      string       `json:"ratio,omitempty"`
	TaskType   string       `json:"task_type,omitempty"`
	Modality   string       `json:"modality,omitempty"`
	Usage      *TaskUsage   `json:"usage,omitempty"`
}

type QueryResponse struct {
	Task VideoTask `json:"task"`
}

type ListResponse struct {
	Items []VideoTask `json:"items"`
	Total int         `json:"total"`
}

type DeleteResponse struct {
	TaskID string `json:"task_id"`
	Action string `json:"action"`
	Status string `json:"status"`
}

type ErrorDetail struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	HTTPCode string `json:"http_code,omitempty"`
}

type ErrorEnvelope struct {
	Type      string      `json:"type"`
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}
