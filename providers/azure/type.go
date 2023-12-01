package azure

import "one-api/types"

type ImageAzureResponse struct {
	ID      string              `json:"id,omitempty"`
	Created int64               `json:"created,omitempty"`
	Expires int64               `json:"expires,omitempty"`
	Result  types.ImageResponse `json:"result,omitempty"`
	Status  string              `json:"status,omitempty"`
	Error   ImageAzureError     `json:"error,omitempty"`
	Header  map[string]string   `json:"header,omitempty"`
}

type ImageAzureError struct {
	Code       string   `json:"code,omitempty"`
	Target     string   `json:"target,omitempty"`
	Message    string   `json:"message,omitempty"`
	Details    []string `json:"details,omitempty"`
	InnerError any      `json:"innererror,omitempty"`
}
