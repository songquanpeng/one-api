package aws

// Request is the request to AWS Llama3
//
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
type Request struct {
	Prompt      string  `json:"prompt"`
	MaxGenLen   int     `json:"max_gen_len,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}

// Response is the response from AWS Llama3
//
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
type Response struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
	StopReason           string `json:"stop_reason"`
}

// {'generation': 'Hi', 'prompt_token_count': 15, 'generation_token_count': 1, 'stop_reason': None}
type StreamResponse struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
	StopReason           string `json:"stop_reason"`
}
