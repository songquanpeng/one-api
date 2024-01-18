package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"one-api/model"

	"github.com/gin-gonic/gin"
)

func RequestJSONConfig() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
}

func GetContext(method, path string, headers map[string]string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	req, _ = http.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func GetGinRouter(method, path string, headers map[string]string, body *io.Reader) *httptest.ResponseRecorder {
	var req *http.Request
	r := gin.Default()

	w := httptest.NewRecorder()
	req, _ = http.NewRequest(method, path, *body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	r.ServeHTTP(w, req)

	return w
}

func GetChannel(channelType int, baseUrl, other, porxy, modelMapping string) model.Channel {
	return model.Channel{
		Type:         channelType,
		BaseURL:      &baseUrl,
		Other:        other,
		Proxy:        porxy,
		ModelMapping: &modelMapping,
		Key:          GetTestToken(),
	}
}
