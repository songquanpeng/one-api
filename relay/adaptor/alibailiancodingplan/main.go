package alibailiancodingplan

import (
	"fmt"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/chat/completions", meta.BaseURL), nil
	default:
		return "", fmt.Errorf("unsupported relay mode %d for ali bailian coding plan", meta.Mode)
	}
}
