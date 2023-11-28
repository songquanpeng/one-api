package providers

import (
	"github.com/gin-gonic/gin"
)

type ClaudeProvider struct {
	ProviderConfig
}

type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func CreateClaudeProvider(c *gin.Context) *ClaudeProvider {
	return &ClaudeProvider{
		ProviderConfig: ProviderConfig{
			BaseURL:         "https://api.anthropic.com",
			ChatCompletions: "/v1/complete",
			Context:         c,
		},
	}
}

// 获取请求头
func (p *ClaudeProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)

	headers["x-api-key"] = p.Context.GetString("api_key")
	headers["Content-Type"] = p.Context.Request.Header.Get("Content-Type")
	headers["Accept"] = p.Context.Request.Header.Get("Accept")
	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json"
	}

	anthropicVersion := p.Context.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	headers["anthropic-version"] = anthropicVersion

	return headers
}

func stopReasonClaude2OpenAI(reason string) string {
	switch reason {
	case "stop_sequence":
		return "stop"
	case "max_tokens":
		return "length"
	default:
		return reason
	}
}
