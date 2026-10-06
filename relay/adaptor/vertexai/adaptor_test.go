package vertexai

import (
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/meta"
)

func TestGetAdaptorSupportsGemini3AndRegionalPrefixes(t *testing.T) {
	models := []string{
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-3.5-flash",
		"au.gemini-3.5-flash",
		"eu.gemini-3.5-pro",
		"gemini-3.1-pro",
	}
	for _, m := range models {
		if adaptor := GetAdaptor(m); adaptor == nil {
			t.Errorf("GetAdaptor(%q) returned nil, expected Gemini adaptor", m)
		}
	}
}

func TestGetRequestURLRegionalGeminiPrefix(t *testing.T) {
	a := &Adaptor{}
	m := &meta.Meta{
		ActualModelName: "au.gemini-3.5-flash",
		IsStream:        true,
		Config: model.ChannelConfig{
			Region:            "australia-southeast1",
			VertexAIProjectID: "test-proj",
		},
	}
	url, err := a.GetRequestURL(m)
	if err != nil {
		t.Fatalf("GetRequestURL failed: %v", err)
	}
	if !strings.HasSuffix(url, ":streamGenerateContent?alt=sse") {
		t.Errorf("expected :streamGenerateContent?alt=sse suffix, got %q", url)
	}

	m.IsStream = false
	url, err = a.GetRequestURL(m)
	if err != nil {
		t.Fatalf("GetRequestURL failed: %v", err)
	}
	if !strings.HasSuffix(url, ":generateContent") {
		t.Errorf("expected :generateContent suffix, got %q", url)
	}
}
