package imagen

type CreateImageRequest struct {
	Instances  []createImageInstance `json:"instances" binding:"required,min=1"`
	Parameters createImageParameters `json:"parameters" binding:"required"`
}

type createImageInstance struct {
	Prompt string `json:"prompt"`
}

type createImageParameters struct {
	SampleCount int `json:"sample_count" binding:"required,min=1"`
}

type CreateImageResponse struct {
	Predictions []createImageResponsePrediction `json:"predictions"`
}

type createImageResponsePrediction struct {
	MimeType           string `json:"mimeType"`
	BytesBase64Encoded string `json:"bytesBase64Encoded"`
}
