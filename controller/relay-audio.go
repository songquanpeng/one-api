package controller

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/dhowden/tag"
	"github.com/gin-gonic/gin"
)

func getMetadata(data []byte) tag.Metadata {
	r := bytes.NewReader(data) // create a bytes.Reader from the data
	m, err := tag.ReadFrom(r)  // read the metadata from the reader
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func relayAudioHelper(c *gin.Context, relayMode int) *OpenAIErrorWithStatusCode {
	// Get form-data file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return errorWrapper(err, "err_get_audio_file", http.StatusBadRequest)
	}
	defer file.Close()

	// Read the file data
	data, err := io.ReadAll(file)
	if err != nil {
		return errorWrapper(err, "err_read_audio_file", http.StatusBadRequest)
	}

	// Get metadata
	m := getMetadata(data)

	// To json

	log.Print(m.Raw())

	return nil
}
