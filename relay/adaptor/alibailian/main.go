package alibailian

import (
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func compatibleBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/compatible-mode/v1") {
		return baseURL
	}
	return baseURL + "/compatible-mode/v1"
}

func GetRequestURL(meta *meta.Meta) (string, error) {
	baseURL := compatibleBaseURL(meta.BaseURL)
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/chat/completions", baseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/embeddings", baseURL), nil
	default:
	}
	return "", fmt.Errorf("unsupported relay mode %d for ali bailian", meta.Mode)
}
