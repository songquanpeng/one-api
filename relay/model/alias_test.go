package model

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/stretchr/testify/assert"
)

func TestResolveModelAlias(t *testing.T) {
	tests := []struct {
		name         string
		standardName string
		channelType  int
		expected     string
	}{
		{
			name:         "OpenRouter gpt-4o alias",
			standardName: "gpt-4o",
			channelType:  channeltype.OpenRouter,
			expected:     "openai/gpt-4o",
		},
		{
			name:         "OpenAI gpt-4o direct",
			standardName: "gpt-4o",
			channelType:  channeltype.OpenAI,
			expected:     "gpt-4o",
		},
		{
			name:         "Anthropic Claude alias",
			standardName: "claude-3-sonnet",
			channelType:  channeltype.Anthropic,
			expected:     "claude-3-sonnet-20240229",
		},
		{
			name:         "Non-existent alias",
			standardName: "non-existent-model",
			channelType:  channeltype.OpenRouter,
			expected:     "non-existent-model",
		},
		{
			name:         "Unsupported channel type",
			standardName: "gpt-4o",
			channelType:  999,
			expected:     "gpt-4o",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveModelAlias(tt.standardName, tt.channelType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegisterModelAlias(t *testing.T) {
	originalMap := ModelAliasMap
	defer func() { ModelAliasMap = originalMap }()

	// Clear the map for testing
	ModelAliasMap = make(map[string]map[int]string)

	RegisterModelAlias("test-model", "actual-test-model", channeltype.OpenRouter)

	assert.Contains(t, ModelAliasMap, "test-model")
	assert.Equal(t, "actual-test-model", ModelAliasMap["test-model"][channeltype.OpenRouter])
}

func TestGetStandardModelName(t *testing.T) {
	tests := []struct {
		name        string
		actualName  string
		channelType int
		expected    string
	}{
		{
			name:        "OpenRouter reverse lookup",
			actualName:  "openai/gpt-4o",
			channelType: channeltype.OpenRouter,
			expected:    "gpt-4o",
		},
		{
			name:        "Anthropic reverse lookup",
			actualName:  "claude-3-sonnet-20240229",
			channelType: channeltype.Anthropic,
			expected:    "claude-3-sonnet",
		},
		{
			name:        "No reverse mapping",
			actualName:  "unknown-model",
			channelType: channeltype.OpenRouter,
			expected:    "unknown-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStandardModelName(tt.actualName, tt.channelType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAliasSupported(t *testing.T) {
	tests := []struct {
		name         string
		standardName string
		channelType  int
		expected     bool
	}{
		{
			name:         "Supported alias",
			standardName: "gpt-4o",
			channelType:  channeltype.OpenRouter,
			expected:     true,
		},
		{
			name:         "Unsupported channel",
			standardName: "gpt-4o",
			channelType:  999,
			expected:     false,
		},
		{
			name:         "Unsupported model",
			standardName: "non-existent",
			channelType:  channeltype.OpenRouter,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAliasSupported(tt.standardName, tt.channelType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSupportedChannelTypes(t *testing.T) {
	channelTypes := GetSupportedChannelTypes("gpt-4o")

	assert.Contains(t, channelTypes, channeltype.OpenRouter)
	assert.Contains(t, channelTypes, channeltype.OpenAI)
	assert.NotContains(t, channelTypes, 999) // non-existent channel type

	emptyTypes := GetSupportedChannelTypes("non-existent-model")
	assert.Empty(t, emptyTypes)
}

func TestGetAllModelAliases(t *testing.T) {
	aliases := GetAllModelAliases()

	assert.NotNil(t, aliases)
	assert.Contains(t, aliases, "gpt-4o")
	assert.Contains(t, aliases["gpt-4o"], channeltype.OpenRouter)
	assert.Equal(t, "openai/gpt-4o", aliases["gpt-4o"][channeltype.OpenRouter])
}

// Benchmark tests
func BenchmarkResolveModelAlias(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ResolveModelAlias("gpt-4o", channeltype.OpenRouter)
	}
}

func BenchmarkGetStandardModelName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetStandardModelName("openai/gpt-4o", channeltype.OpenRouter)
	}
}
