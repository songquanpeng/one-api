package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
)

func TestPinstripeGetFullRequestURL(t *testing.T) {
	baseURL := "https://pinstripes.io"
	requestURL := "/v1/chat/completions"
	got := GetFullRequestURL(baseURL, requestURL, channeltype.Pinstripes)
	want := "https://pinstripes.io/v1/chat/completions"
	if got != want {
		t.Errorf("GetFullRequestURL() = %q, want %q", got, want)
	}
}
