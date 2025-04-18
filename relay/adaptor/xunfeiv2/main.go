package xunfeiv2

import (
	"fmt"

	"github.com/songquanpeng/one-api/relay/meta"
)

func GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.ActualModelName {
	case "x1":
		return fmt.Sprintf("%s/v2/chat/completions", meta.BaseURL), nil
	default:
	}
	return fmt.Sprintf("%s/v2/chat/completions", meta.BaseURL), nil
}
