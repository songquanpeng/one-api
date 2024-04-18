package drives

import (
	"bytes"
	"fmt"
	"net/http"
	"one-api/common/requester"
)

var imgurUploadURL = "https://api.imgur.com/3/image"

type ImgurUpload struct {
	ClientId string
}

func NewImgurUpload(clientId string) *ImgurUpload {
	return &ImgurUpload{
		ClientId: clientId,
	}
}

type ImgurResponse struct {
	Status  int       `json:"status"`
	Success bool      `json:"success"`
	Data    ImgurData `json:"data,omitempty"`
}

type ImgurData struct {
	Link string `json:"link"`
}

func (i *ImgurUpload) Name() string {
	return "Imgur"
}

func (i *ImgurUpload) Upload(data []byte, fileName string) (string, error) {
	client := requester.NewHTTPRequester("", nil)

	var formBody bytes.Buffer
	builder := client.CreateFormBuilder(&formBody)

	err := builder.CreateFormFileReader("image", bytes.NewReader(data), fileName)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	builder.Close()

	headers := map[string]string{
		"Authorization": "Client-ID " + i.ClientId,
	}

	req, err := client.NewRequest(
		http.MethodPost,
		imgurUploadURL,
		client.WithBody(&formBody),
		client.WithHeader(headers),
		client.WithContentType(builder.FormDataContentType()))
	req.ContentLength = int64(formBody.Len())

	if err != nil {
		return "", fmt.Errorf("new request failed: %w", err)
	}

	defer req.Body.Close()

	imgurResponse := &ImgurResponse{}
	_, errWithCode := client.SendRequest(req, imgurResponse, false)
	if errWithCode != nil {
		return "", fmt.Errorf("%s", errWithCode.Message)
	}

	if !imgurResponse.Success {
		return "", fmt.Errorf("upload failed Status: %d", imgurResponse.Status)
	}

	return imgurResponse.Data.Link, nil
}
