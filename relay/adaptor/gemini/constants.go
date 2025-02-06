package gemini

// https://ai.google.dev/models/gemini

var ModelList = []string{
	"gemini-pro", "gemini-1.0-pro",
	// "gemma-2-2b-it", "gemma-2-9b-it", "gemma-2-27b-it",
	"gemini-1.5-flash", "gemini-1.5-flash-8b",
	"gemini-1.5-pro", "gemini-1.5-pro-experimental",
	"text-embedding-004", "aqa",
	"gemini-2.0-flash", "gemini-2.0-flash-exp",
	"gemini-2.0-flash-lite-preview-02-05",
	"gemini-2.0-flash-thinking-exp-01-21",
	"gemini-2.0-pro-exp-02-05",
}

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
	for _, m := range ModelsSupportSystemInstruction {
		if m == model {
			return true
		}
	}

	return false
}
