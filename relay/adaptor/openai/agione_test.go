package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
)

func TestGetFullRequestURLForAGIone(t *testing.T) {
	tests := []struct {
		name       string
		baseURL    string
		requestURL string
		want       string
	}{
		{
			name:       "default versioned base URL",
			baseURL:    "https://agione.pro/hyperone/xapi/api/v1",
			requestURL: "/v1/chat/completions",
			want:       "https://agione.pro/hyperone/xapi/api/v1/chat/completions",
		},
		{
			name:       "default versioned base URL with trailing slash",
			baseURL:    "https://agione.pro/hyperone/xapi/api/v1/",
			requestURL: "/v1/chat/completions",
			want:       "https://agione.pro/hyperone/xapi/api/v1/chat/completions",
		},
		{
			name:       "unversioned base URL",
			baseURL:    "https://agione.pro/hyperone/xapi/api",
			requestURL: "/v1/chat/completions",
			want:       "https://agione.pro/hyperone/xapi/api/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFullRequestURL(tt.baseURL, tt.requestURL, channeltype.AGIone)
			if got != tt.want {
				t.Fatalf("GetFullRequestURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
