package minimax

import (
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	if meta.Mode == relaymode.ChatCompletions {
		return fmt.Sprintf("%s/v1/text/chatcompletion_v2", meta.BaseURL), nil
	}
	if meta.Mode == relaymode.AudioSpeech {
		baseURL := strings.TrimSuffix(strings.TrimRight(meta.BaseURL, "/"), "/v1")
		if strings.Contains(baseURL, "api.minimax.chat") {
			baseURL = strings.Replace(baseURL, "api.minimax.chat", "api.minimax.io", 1)
		}
		return fmt.Sprintf("%s/v1/t2a_v2", baseURL), nil
	}
	return "", fmt.Errorf("unsupported relay mode %d for minimax", meta.Mode)
}
