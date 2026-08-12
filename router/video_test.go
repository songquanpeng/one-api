package router

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVideoRoutesAreRegisteredWithoutNativeV2Ingress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetRelayRouter(engine)

	wanted := map[string]bool{
		"POST /v1/videos":                  false,
		"GET /v1/videos":                   false,
		"GET /v1/videos/:video_id":         false,
		"GET /v1/videos/:video_id/content": false,
		"DELETE /v1/videos/:video_id":      false,
	}
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := wanted[key]; exists {
			wanted[key] = true
		}
		if route.Path == "/v2/video_generation" || route.Path == "/v2/query/video_generation" || route.Path == "/v2/h3_context_ir" {
			t.Fatalf("native SiftQ route must not be public: %s", key)
		}
	}
	for route, found := range wanted {
		if !found {
			t.Fatalf("route not registered: %s", route)
		}
	}
}
