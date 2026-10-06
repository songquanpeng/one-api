package gemini

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func strPtr(s string) *string {
	return &s
}

func TestExtractGeminiBaseModelAndSystemInstruction(t *testing.T) {
	cases := []struct {
		model       string
		wantBase    string
		wantSysInst bool
	}{
		{"gemini-3.5-flash", "gemini-3.5-flash", true},
		{"au.gemini-3.5-flash", "gemini-3.5-flash", true},
		{"eu.gemini-3.5-pro", "gemini-3.5-pro", true},
		{"models/gemini-3.1-pro", "gemini-3.1-pro", true},
		{"gemini-2.5-flash", "gemini-2.5-flash", true},
		{"gemini-1.5-pro", "gemini-1.5-pro", true},
		{"gemini-1.0-pro", "gemini-1.0-pro", false},
		{"gemini-pro", "gemini-pro", false},
	}
	for _, tc := range cases {
		if got := ExtractGeminiBaseModel(tc.model); got != tc.wantBase {
			t.Errorf("ExtractGeminiBaseModel(%q) = %q, want %q", tc.model, got, tc.wantBase)
		}
		if got := IsModelSupportSystemInstruction(tc.model); got != tc.wantSysInst {
			t.Errorf("IsModelSupportSystemInstruction(%q) = %v, want %v", tc.model, got, tc.wantSysInst)
		}
	}
}

func TestConvertRequestThinkingConfig(t *testing.T) {
	cases := []struct {
		name         string
		model        string
		effort       *string
		wantNil      bool
		wantThoughts bool
		wantLevel    string
		wantBudget   *int
	}{
		{
			name:         "gemini-3.5-flash none",
			model:        "gemini-3.5-flash",
			effort:       strPtr("none"),
			wantThoughts: false,
			wantLevel:    "minimal",
		},
		{
			name:         "au.gemini-3.5-flash none",
			model:        "au.gemini-3.5-flash",
			effort:       strPtr("none"),
			wantThoughts: false,
			wantLevel:    "minimal",
		},
		{
			name:         "au.gemini-3.5-flash minimal",
			model:        "au.gemini-3.5-flash",
			effort:       strPtr("minimal"),
			wantThoughts: true,
			wantLevel:    "minimal",
		},
		{
			name:         "gemini-3.5-pro none",
			model:        "gemini-3.5-pro",
			effort:       strPtr("none"),
			wantThoughts: false,
			wantLevel:    "low",
		},
		{
			name:         "gemini-3.5-pro minimal",
			model:        "gemini-3.5-pro",
			effort:       strPtr("minimal"),
			wantThoughts: true,
			wantLevel:    "low",
		},
		{
			name:         "gemini-3.1-pro minimal",
			model:        "gemini-3.1-pro",
			effort:       strPtr("minimal"),
			wantThoughts: true,
			wantLevel:    "low",
		},
		{
			name:         "gemini-2.5-flash none",
			model:        "gemini-2.5-flash",
			effort:       strPtr("none"),
			wantThoughts: false,
			wantBudget:   func() *int { v := 0; return &v }(),
		},
		{
			name:         "gemini-2.5-pro none",
			model:        "gemini-2.5-pro",
			effort:       strPtr("none"),
			wantThoughts: false,
			wantBudget:   func() *int { v := 128; return &v }(),
		},
		{
			name:         "gemini-2.5-flash low",
			model:        "gemini-2.5-flash",
			effort:       strPtr("low"),
			wantThoughts: true,
			wantBudget:   func() *int { v := 1024; return &v }(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := ConvertRequest(model.GeneralOpenAIRequest{
				Model:           tc.model,
				ReasoningEffort: tc.effort,
				Messages: []model.Message{
					{Role: "system", Content: "Be concise."},
					{Role: "user", Content: "Hi"},
				},
			})
			if req.SystemInstruction == nil || len(req.SystemInstruction.Parts) != 1 || req.SystemInstruction.Parts[0].Text != "Be concise." {
				t.Fatalf("expected system_instruction to be populated, got %+v", req.SystemInstruction)
			}
			cfg := req.GenerationConfig.ThinkingConfig
			if tc.wantNil {
				if cfg != nil {
					t.Fatalf("expected nil ThinkingConfig, got %+v", cfg)
				}
				return
			}
			if cfg == nil {
				t.Fatalf("expected non-nil ThinkingConfig")
			}
			if cfg.IncludeThoughts == nil || *cfg.IncludeThoughts != tc.wantThoughts {
				t.Errorf("IncludeThoughts = %v, want %v", cfg.IncludeThoughts, tc.wantThoughts)
			}
			if cfg.ThinkingLevel != tc.wantLevel {
				t.Errorf("ThinkingLevel = %q, want %q", cfg.ThinkingLevel, tc.wantLevel)
			}
			if tc.wantBudget == nil {
				if cfg.ThinkingBudget != nil {
					t.Errorf("ThinkingBudget = %v, want nil", *cfg.ThinkingBudget)
				}
			} else {
				if cfg.ThinkingBudget == nil || *cfg.ThinkingBudget != *tc.wantBudget {
					t.Errorf("ThinkingBudget = %v, want %d", cfg.ThinkingBudget, *tc.wantBudget)
				}
			}
		})
	}
}

func TestResponseThoughtSeparationAndUsageMetadata(t *testing.T) {
	chatResp := &ChatResponse{
		Candidates: []ChatCandidate{
			{
				Content: ChatContent{
					Role: "model",
					Parts: []Part{
						{Text: "Step 1: analyze", Thought: true},
						{Text: "42", Thought: false},
					},
				},
				FinishReason: "STOP",
			},
		},
		UsageMetadata: &UsageMetadata{
			PromptTokenCount:        2048,
			CachedContentTokenCount: 1536,
			CandidatesTokenCount:    2,
			ThoughtsTokenCount:      350,
			TotalTokenCount:         2400,
		},
	}

	openAIResp := responseGeminiChat2OpenAI(chatResp)
	if len(openAIResp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(openAIResp.Choices))
	}
	if openAIResp.Choices[0].Message.Content != "42" {
		t.Errorf("Content = %v, want 42", openAIResp.Choices[0].Message.Content)
	}
	if openAIResp.Choices[0].Message.ReasoningContent != "Step 1: analyze" {
		t.Errorf("ReasoningContent = %v, want 'Step 1: analyze'", openAIResp.Choices[0].Message.ReasoningContent)
	}

	streamResp := streamResponseGeminiChat2OpenAI(chatResp)
	if streamResp.Choices[0].Delta.Content != "42" {
		t.Errorf("Delta.Content = %v, want 42", streamResp.Choices[0].Delta.Content)
	}
	if streamResp.Choices[0].Delta.ReasoningContent != "Step 1: analyze" {
		t.Errorf("Delta.ReasoningContent = %v, want 'Step 1: analyze'", streamResp.Choices[0].Delta.ReasoningContent)
	}
	if streamResp.Usage == nil {
		t.Fatalf("expected streamResp.Usage to be populated")
	}
	if streamResp.Usage.PromptTokens != 2048 || streamResp.Usage.CompletionTokens != 352 || streamResp.Usage.TotalTokens != 2400 {
		t.Errorf("unexpected Usage counts: %+v", streamResp.Usage)
	}
	if streamResp.Usage.PromptTokensDetails == nil || streamResp.Usage.PromptTokensDetails.CachedTokens != 1536 {
		t.Errorf("expected CachedTokens=1536, got %+v", streamResp.Usage.PromptTokensDetails)
	}
	if streamResp.Usage.CompletionTokensDetails == nil || streamResp.Usage.CompletionTokensDetails.ReasoningTokens != 350 {
		t.Errorf("expected ReasoningTokens=350, got %+v", streamResp.Usage.CompletionTokensDetails)
	}
}

func TestResponseParallelToolCallsAndFinishReason(t *testing.T) {
	chatResp := &ChatResponse{
		Candidates: []ChatCandidate{
			{
				Content: ChatContent{
					Role: "model",
					Parts: []Part{
						{
							FunctionCall: &FunctionCall{
								FunctionName: "get_weather",
								Arguments:    map[string]any{"city": "Sydney"},
							},
						},
						{
							FunctionCall: &FunctionCall{
								FunctionName: "get_weather",
								Arguments:    map[string]any{"city": "Melbourne"},
							},
						},
					},
				},
				FinishReason: "STOP",
			},
		},
	}

	openAIResp := responseGeminiChat2OpenAI(chatResp)
	if len(openAIResp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(openAIResp.Choices))
	}
	choice := openAIResp.Choices[0]
	if len(choice.Message.ToolCalls) != 2 {
		t.Fatalf("expected 2 parallel tool calls, got %d", len(choice.Message.ToolCalls))
	}
	if choice.FinishReason != "tool_calls" {
		t.Errorf("FinishReason = %q, want %q", choice.FinishReason, "tool_calls")
	}
}
