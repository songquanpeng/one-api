package replicate

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error { return nil }

// Custom FileHeader to override Open method
type customFileHeader struct {
	*multipart.FileHeader
	openFunc func() (multipart.File, error)
}

func (c *customFileHeader) Open() (multipart.File, error) {
	return c.openFunc()
}

func TestOpenaiImageEditRequest_toFluxRemixRequest(t *testing.T) {
	// Create a simple image for testing
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: image.Black}, image.Point{}, draw.Src)
	var imgBuf bytes.Buffer
	err := png.Encode(&imgBuf, img)
	require.NoError(t, err)

	// Create a simple mask for testing
	mask := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(mask, mask.Bounds(), &image.Uniform{C: image.Black}, image.Point{}, draw.Src)
	var maskBuf bytes.Buffer
	err = png.Encode(&maskBuf, mask)
	require.NoError(t, err)

	// Create a multipart.FileHeader from the image and mask bytes
	imgFileHeader, err := createFileHeader("image", "test.png", imgBuf.Bytes())
	require.NoError(t, err)
	maskFileHeader, err := createFileHeader("mask", "test.png", maskBuf.Bytes())
	require.NoError(t, err)

	req := &OpenaiImageEditRequest{
		Image:          imgFileHeader,
		Mask:           maskFileHeader,
		Prompt:         "Test prompt",
		Model:          "test-model",
		ResponseFormat: "b64_json",
	}

	fluxReq, err := req.toFluxRemixRequest()
	require.NoError(t, err)
	require.NotNil(t, fluxReq)
	require.Equal(t, req.Prompt, fluxReq.Input.Prompt)
	require.NotEmpty(t, fluxReq.Input.Image)
	require.NotEmpty(t, fluxReq.Input.Mask)
}

// createFileHeader creates a multipart.FileHeader from file bytes
func createFileHeader(fieldname, filename string, fileBytes []byte) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file field
	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return nil, err
	}

	// Write the file bytes to the form file field
	_, err = part.Write(fileBytes)
	if err != nil {
		return nil, err
	}

	// Close the writer to finalize the form
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Parse the multipart form
	req := &http.Request{
		Header: http.Header{},
		Body:   io.NopCloser(body),
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	err = req.ParseMultipartForm(int64(body.Len()))
	if err != nil {
		return nil, err
	}

	// Retrieve the file header from the parsed form
	fileHeader := req.MultipartForm.File[fieldname][0]
	return fileHeader, nil
}
