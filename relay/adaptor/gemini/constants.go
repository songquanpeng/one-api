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
	// "gemini-1.0-pro-002",
	// "gemini-1.5-flash", "gemini-1.5-flash-001", "gemini-1.5-flash-002",
	// "gemini-1.5-flash-8b",
	// "gemini-1.5-pro", "gemini-1.5-pro-001", "gemini-1.5-pro-002",
	// "gemini-1.5-pro-experimental",
	"gemini-2.0-flash", "gemini-2.0-flash-exp",
	"gemini-2.0-flash-thinking-exp-01-21",
}

// IsModelSupportSystemInstruction check if the model support system instruction.
//
// Because the main version of Go is 1.20, slice.Contains cannot be used
func IsModelSupportSystemInstruction(model string) bool {
	if strings.HasPrefix(model, "gemini-1.5-") || strings.HasPrefix(model, "gemini-2.") || strings.HasPrefix(model, "gemini-3.") {
		return true
	}
	for _, m := range ModelsSupportSystemInstruction {
		if m == model {
			return true
		}
	}

	return false
}
