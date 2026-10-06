package gemini

import (
	"strings"

	"github.com/songquanpeng/one-api/relay/adaptor/geminiv2"
)

// https://ai.google.dev/models/gemini

var ModelList = geminiv2.ModelList

// ModelsSupportSystemInstruction is the list of models that support system instruction.
//
// https://cloud.google.com/vertex-ai/generative-ai/docs/learn/prompts/system-instructions
var ModelsSupportSystemInstruction = []string{
	"gemini-2.0-flash", "gemini-2.0-flash-exp",
	"gemini-2.0-flash-thinking-exp-01-21",
}

// ExtractGeminiBaseModel strips resource path prefixes (e.g. "models/") and
// regional endpoint prefixes (e.g. "au.gemini-3.5-flash") to return the canonical
// lowercase Gemini model name.
func ExtractGeminiBaseModel(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	if slash := strings.LastIndex(m, "/"); slash >= 0 {
		m = m[slash+1:]
	}
	if idx := strings.Index(m, "gemini-"); idx >= 0 {
		return m[idx:]
	}
	return m
}

// IsModelSupportSystemInstruction checks if the model supports system_instruction.
// All Gemini 1.5, 2.x, and 3.x+ models support system_instruction.
func IsModelSupportSystemInstruction(model string) bool {
	for _, m := range ModelsSupportSystemInstruction {
		if m == model {
			return true
		}
	}
	base := ExtractGeminiBaseModel(model)
	if base == "gemini-1.0-pro" || strings.HasPrefix(base, "gemini-1.0-pro-") || base == "gemini-pro" || base == "gemini-pro-vision" {
		return false
	}
	return strings.HasPrefix(base, "gemini-")
}
