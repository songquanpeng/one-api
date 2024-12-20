package replicate

import (
	"time"

	"github.com/pkg/errors"
)

// DrawImageRequest draw image by fluxpro
//
// https://replicate.com/black-forest-labs/flux-pro?prediction=kg1krwsdf9rg80ch1sgsrgq7h8&output=json
type DrawImageRequest struct {
	Input ImageInput `json:"input"`
}

// ImageInput is input of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-1.1-pro/api/schema
type ImageInput struct {
	Steps           int    `json:"steps" binding:"required,min=1"`
	Prompt          string `json:"prompt" binding:"required,min=5"`
	ImagePrompt     string `json:"image_prompt"`
	Guidance        int    `json:"guidance" binding:"required,min=2,max=5"`
	Interval        int    `json:"interval" binding:"required,min=1,max=4"`
	AspectRatio     string `json:"aspect_ratio" binding:"required,oneof=1:1 16:9 2:3 3:2 4:5 5:4 9:16"`
	SafetyTolerance int    `json:"safety_tolerance" binding:"required,min=1,max=5"`
	Seed            int    `json:"seed"`
	NImages         int    `json:"n_images" binding:"required,min=1,max=8"`
	Width           int    `json:"width" binding:"required,min=256,max=1440"`
	Height          int    `json:"height" binding:"required,min=256,max=1440"`
}

// InpaintingImageByFlusReplicateRequest is request to inpainting image by flux pro
//
// https://replicate.com/black-forest-labs/flux-fill-pro/api/schema
type InpaintingImageByFlusReplicateRequest struct {
	Input FluxInpaintingInput `json:"input"`
}

// FluxInpaintingInput is input of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-fill-pro/api/schema
type FluxInpaintingInput struct {
	Mask             string `json:"mask" binding:"required"`
	Image            string `json:"image" binding:"required"`
	Seed             int    `json:"seed"`
	Steps            int    `json:"steps" binding:"required,min=1"`
	Prompt           string `json:"prompt" binding:"required,min=5"`
	Guidance         int    `json:"guidance" binding:"required,min=2,max=5"`
	OutputFormat     string `json:"output_format"`
	SafetyTolerance  int    `json:"safety_tolerance" binding:"required,min=1,max=5"`
	PromptUnsampling bool   `json:"prompt_unsampling"`
}

// ImageResponse is response of DrawImageByFluxProRequest
//
// https://replicate.com/black-forest-labs/flux-pro?prediction=kg1krwsdf9rg80ch1sgsrgq7h8&output=json
type ImageResponse struct {
	CompletedAt time.Time        `json:"completed_at"`
	CreatedAt   time.Time        `json:"created_at"`
	DataRemoved bool             `json:"data_removed"`
	Error       string           `json:"error"`
	ID          string           `json:"id"`
	Input       DrawImageRequest `json:"input"`
	Logs        string           `json:"logs"`
	Metrics     FluxMetrics      `json:"metrics"`
	// Output could be `string` or `[]string`
	Output    any       `json:"output"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"`
	URLs      FluxURLs  `json:"urls"`
	Version   string    `json:"version"`
}

func (r *ImageResponse) GetOutput() ([]string, error) {
	switch v := r.Output.(type) {
	case string:
		return []string{v}, nil
	case []string:
		return v, nil
	case nil:
		return nil, nil
	case []interface{}:
		// convert []interface{} to []string
		ret := make([]string, len(v))
		for idx, vv := range v {
			if vvv, ok := vv.(string); ok {
				ret[idx] = vvv
			} else {
				return nil, errors.Errorf("unknown output type: [%T]%v", vv, vv)
			}
		}

		return ret, nil
	default:
		return nil, errors.Errorf("unknown output type: [%T]%v", r.Output, r.Output)
	}
}

// FluxMetrics is metrics of ImageResponse
type FluxMetrics struct {
	ImageCount  int     `json:"image_count"`
	PredictTime float64 `json:"predict_time"`
	TotalTime   float64 `json:"total_time"`
}

// FluxURLs is urls of ImageResponse
type FluxURLs struct {
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}

type ReplicateChatRequest struct {
	Input ChatInput `json:"input" form:"input" binding:"required"`
}

// ChatInput is input of ChatByReplicateRequest
//
// https://replicate.com/meta/meta-llama-3.1-405b-instruct/api/schema
type ChatInput struct {
	TopK             int     `json:"top_k"`
	TopP             float64 `json:"top_p"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	MinTokens        int     `json:"min_tokens"`
	Temperature      float64 `json:"temperature"`
	SystemPrompt     string  `json:"system_prompt"`
	StopSequences    string  `json:"stop_sequences"`
	PromptTemplate   string  `json:"prompt_template"`
	PresencePenalty  float64 `json:"presence_penalty"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
}

// ChatResponse is response of ChatByReplicateRequest
//
// https://replicate.com/meta/meta-llama-3.1-405b-instruct/examples?input=http&output=json
type ChatResponse struct {
	CompletedAt time.Time   `json:"completed_at"`
	CreatedAt   time.Time   `json:"created_at"`
	DataRemoved bool        `json:"data_removed"`
	Error       string      `json:"error"`
	ID          string      `json:"id"`
	Input       ChatInput   `json:"input"`
	Logs        string      `json:"logs"`
	Metrics     FluxMetrics `json:"metrics"`
	// Output could be `string` or `[]string`
	Output    []string        `json:"output"`
	StartedAt time.Time       `json:"started_at"`
	Status    string          `json:"status"`
	URLs      ChatResponseUrl `json:"urls"`
	Version   string          `json:"version"`
}

// ChatResponseUrl is task urls of ChatResponse
type ChatResponseUrl struct {
	Stream string `json:"stream"`
	Get    string `json:"get"`
	Cancel string `json:"cancel"`
}
