package drives

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common/requester"
)

var smUploadURL = "https://sm.ms/api/v2/upload"

type SMUpload struct {
	Secret string
}

type SMData struct {
	URL string `json:"url"`
	// FileID    int    `json:"file_id"`
	// Width     int    `json:"width"`
	// Height    int    `json:"height"`
	// Filename  string `json:"filename"`
	// Storename string `json:"storename"`
	// Size      int    `json:"size"`
	// Path      string `json:"path"`
	// Hash      string `json:"hash"`
	// Delete    string `json:"delete"`
	// Page      string `json:"page"`
}

type SMResponse struct {
	Success   bool   `json:"success"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      SMData `json:"data"`
	RequestID string `json:"RequestId"`
}

func NewSMUpload(secret string) *SMUpload {
	return &SMUpload{
		Secret: secret,
	}
}

func (sm *SMUpload) Name() string {
	return "SM.MS"
}

func (sm *SMUpload) Upload(data []byte, fileName string) (string, error) {
	client := requester.NewHTTPRequester("", nil)

	var formBody bytes.Buffer
	builder := client.CreateFormBuilder(&formBody)

	err := builder.CreateFormFileReader("smfile", bytes.NewReader(data), fileName)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	builder.WriteField("format", "json")

	headers := map[string]string{
		"Content-type":  "application/json",
		"Authorization": sm.Secret,
	}

	req, err := client.NewRequest(
		http.MethodPost,
		smUploadURL,
		client.WithBody(&formBody),
		client.WithHeader(headers),
		client.WithContentType(builder.FormDataContentType()))
	req.ContentLength = int64(formBody.Len())

	if err != nil {
		return "", fmt.Errorf("new request failed: %w", err)
	}

	defer req.Body.Close()

	smResponse := &SMResponse{}
	_, errWithCode := client.SendRequest(req, smResponse, false)
	if errWithCode != nil {
		return "", fmt.Errorf("%s", errWithCode.Message)
	}

	if !smResponse.Success {
		return "", fmt.Errorf("upload failed: %s", smResponse.Message)
	}

	return smResponse.Data.URL, nil
}
