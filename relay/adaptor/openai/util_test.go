package openai

import (
	"strings"
	"testing"
)

func TestExtractThinkContent(t *testing.T) {
	tests := []struct {
		name                 string
		input                string
		wantCleanContent     string
		wantReasoningContent string
	}{
		{
			name:                 "No think tag",
			input:                "Hello, world!",
			wantCleanContent:     "Hello, world!",
			wantReasoningContent: "",
		},
		{
			name:                 "Empty input",
			input:                "",
			wantCleanContent:     "",
			wantReasoningContent: "",
		},
		{
			name:                 "Single think tag",
			input:                "Hello, <think>This is reasoning</think> world!",
			wantCleanContent:     "Hello, world!",
			wantReasoningContent: "This is reasoning\n",
		},
		{
			name:                 "Multiple think tags",
			input:                "<think>First reasoning</think>Hello, <think>Second reasoning</think> world!",
			wantCleanContent:     "Hello, world!",
			wantReasoningContent: "First reasoning\nSecond reasoning\n",
		},
		{
			name:                 "Think tag with newlines",
			input:                "Hello, <think>This is\nmulti-line\nreasoning</think> world!",
			wantCleanContent:     "Hello, world!",
			wantReasoningContent: "This is\nmulti-line\nreasoning\n",
		},
		{
			name:                 "Only think tag",
			input:                "<think>Only reasoning</think>",
			wantCleanContent:     "",
			wantReasoningContent: "Only reasoning\n",
		},
		{
			name:                 "Incomplete think tag (should be ignored)",
			input:                "Hello, <think>Incomplete reasoning world!",
			wantCleanContent:     "Hello, <think>Incomplete reasoning world!",
			wantReasoningContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCleanContent, gotReasoningContent := extractThinkContent(tt.input)
			if gotCleanContent != tt.wantCleanContent {
				t.Errorf("extractThinkContent() gotCleanContent = %q, want %q", gotCleanContent, tt.wantCleanContent)
			}
			if gotReasoningContent != tt.wantReasoningContent {
				t.Errorf("extractThinkContent() gotReasoningContent = %q, want %q", gotReasoningContent, tt.wantReasoningContent)
			}
		})
	}
}

func TestProcessStreamThinkTag(t *testing.T) {
	tests := []struct {
		name                 string
		content              string
		initialInThinkTag    bool
		wantCleanContent     string
		wantReasoningContent string
		wantStillInThinkTag  bool
	}{
		{
			name:                 "Empty content",
			content:              "",
			initialInThinkTag:    false,
			wantCleanContent:     "",
			wantReasoningContent: "",
			wantStillInThinkTag:  false,
		},
		{
			name:                 "Regular content",
			content:              "Hello, world!",
			initialInThinkTag:    false,
			wantCleanContent:     "Hello, world!",
			wantReasoningContent: "",
			wantStillInThinkTag:  false,
		},
		{
			name:                 "Content with <think> tag",
			content:              "Hello, <think>reasoning",
			initialInThinkTag:    false,
			wantCleanContent:     "Hello, ",
			wantReasoningContent: "reasoning",
			wantStillInThinkTag:  true,
		},
		{
			name:                 "Content with </think> tag",
			content:              "reasoning</think> world!",
			initialInThinkTag:    true,
			wantCleanContent:     " world!",
			wantReasoningContent: "reasoning",
			wantStillInThinkTag:  false,
		},
		{
			name:                 "Content inside <think> tag",
			content:              "reasoning content",
			initialInThinkTag:    true,
			wantCleanContent:     "",
			wantReasoningContent: "reasoning content",
			wantStillInThinkTag:  true,
		},
		{
			name:                 "Content with both <think> and </think> tags",
			content:              "Hello, <think>reasoning</think> world!",
			initialInThinkTag:    false,
			wantCleanContent:     "Hello, world!",
			wantReasoningContent: "reasoning",
			wantStillInThinkTag:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reasoningBuilder strings.Builder
			gotCleanContent, gotReasoningContent, gotStillInThinkTag := processStreamThinkTag(tt.content, tt.initialInThinkTag, &reasoningBuilder)
			
			if gotCleanContent != tt.wantCleanContent {
				t.Errorf("processStreamThinkTag() gotCleanContent = %q, want %q", gotCleanContent, tt.wantCleanContent)
			}
			if gotReasoningContent != tt.wantReasoningContent {
				t.Errorf("processStreamThinkTag() gotReasoningContent = %q, want %q", gotReasoningContent, tt.wantReasoningContent)
			}
			if gotStillInThinkTag != tt.wantStillInThinkTag {
				t.Errorf("processStreamThinkTag() gotStillInThinkTag = %v, want %v", gotStillInThinkTag, tt.wantStillInThinkTag)
			}
		})
	}
}
