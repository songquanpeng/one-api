package replicate

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToFluxRemixRequest(t *testing.T) {
	// Prepare input data
	imageData := []byte{0x89, 0x50, 0x4E, 0x47} // Simulates PNG magic bytes
	maskData := []byte{
		0, 0, 0, 0, // Transparent pixel
		255, 255, 255, 255, // Opaque white pixel
	}
	prompt := "Test prompt"
	model := "Test model"
	responseType := "json"

	// convert image and mask to FileHeader
	imageFileHeader, err := createFileHeader("image", "image.png", imageData)
	require.NoError(t, err)
	maskFileHeader, err := createFileHeader("mask", "mask.png", maskData)
	require.NoError(t, err)

	request := OpenaiImageEditRequest{
		Image:          imageFileHeader,
		Mask:           maskFileHeader,
		Prompt:         prompt,
		Model:          model,
		ResponseFormat: responseType,
	}

	// Call the method under test
	fluxRequest, err := request.toFluxRemixRequest()
	require.NoError(t, err)

	// Verify FluxInpaintingInput fields
	require.NotNil(t, fluxRequest)
	require.Equal(t, prompt, fluxRequest.Input.Prompt)
	require.Equal(t, 30, fluxRequest.Input.Steps)
	require.Equal(t, 3, fluxRequest.Input.Guidance)
	require.Equal(t, 5, fluxRequest.Input.SafetyTolerance)
	require.False(t, fluxRequest.Input.PromptUnsampling)

	// Check image field (Base64 encoded)
	expectedImageBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageData)
	require.Equal(t, expectedImageBase64, fluxRequest.Input.Image)

	// Check mask field (Base64 encoded and inverted transparency)
	expectedInvertedMask := []byte{
		255, 255, 255, 255, // Transparent pixel inverted to black
		255, 255, 255, 255, // Opaque white pixel remains the same
	}
	expectedMaskBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(expectedInvertedMask)
	require.Equal(t, expectedMaskBase64, fluxRequest.Input.Mask)

	// Verify seed
	// Since the seed is generated based on the current time, we validate its presence
	require.NotZero(t, fluxRequest.Input.Seed)
	require.True(t, fluxRequest.Input.Seed > 0)

	// Additional assertions can be added as necessary
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
