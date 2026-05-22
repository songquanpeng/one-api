package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRequestModelDefaultsVideoGenerationsModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", strings.NewReader(`{"prompt":"make a video"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	model, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("get request model: %v", err)
	}
	if model != "fake-seedance-lite" {
		t.Fatalf("model = %q, want fake-seedance-lite", model)
	}
}

func TestGetRequestModelPreservesExplicitVideoGenerationsModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos/generations", strings.NewReader(`{"model":"custom-video-model","prompt":"make a video"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	model, err := getRequestModel(c)
	if err != nil {
		t.Fatalf("get request model: %v", err)
	}
	if model != "custom-video-model" {
		t.Fatalf("model = %q, want custom-video-model", model)
	}
}
