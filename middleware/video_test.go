package middleware

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVideoCreateModelExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jsonContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	jsonContext.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewBufferString(`{"prompt":"hello","model":"MiniMax-H3"}`))
	jsonContext.Request.Header.Set("Content-Type", "application/json")
	modelName, err := getRequestModel(jsonContext)
	if err != nil || modelName != "MiniMax-H3" {
		t.Fatalf("JSON model extraction failed: model=%q err=%v", modelName, err)
	}
	if !shouldCheckModel(jsonContext) {
		t.Fatal("video create must enforce token model permissions")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("prompt", "hello")
	_ = writer.WriteField("model", "MiniMax-H3")
	_ = writer.Close()
	multipartContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	multipartContext.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", &body)
	multipartContext.Request.Header.Set("Content-Type", writer.FormDataContentType())
	modelName, err = getRequestModel(multipartContext)
	if err != nil || modelName != "MiniMax-H3" {
		t.Fatalf("multipart model extraction failed: model=%q err=%v", modelName, err)
	}

	getContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	getContext.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/video-1", nil)
	if shouldCheckModel(getContext) {
		t.Fatal("video retrieval must use the persisted channel binding, not a request model")
	}
}
