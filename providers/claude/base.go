package claude

import (
	"one-api/providers/base"

	"github.com/gin-gonic/gin"
)

type ClaudeProvider struct {
	base.BaseProvider
}

func CreateClaudeProvider(c *gin.Context) *ClaudeProvider {
	return &ClaudeProvider{
		BaseProvider: base.BaseProvider{
			BaseURL:         "https://api.anthropic.com",
			ChatCompletions: "/v1/complete",
			Context:         c,
		},
	}
}

// 获取请求头
func (p *ClaudeProvider) GetRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	p.CommonRequestHeaders(headers)

	headers["x-api-key"] = p.Context.GetString("api_key")
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
