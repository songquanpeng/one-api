package openai

var ModelList = []string{
	// GPT-3.5 series
	"gpt-3.5-turbo", "gpt-3.5-turbo-0301", "gpt-3.5-turbo-0613", "gpt-3.5-turbo-1106", "gpt-3.5-turbo-0125",
	"gpt-3.5-turbo-16k", "gpt-3.5-turbo-16k-0613",
	"gpt-3.5-turbo-instruct",
	// GPT-4 series
	"gpt-4", "gpt-4-0314", "gpt-4-0613", "gpt-4-1106-preview", "gpt-4-0125-preview",
	"gpt-4-32k", "gpt-4-32k-0314", "gpt-4-32k-0613",
	"gpt-4-turbo-preview", "gpt-4-turbo", "gpt-4-turbo-2024-04-09",
	"gpt-4-vision-preview",
	// GPT-4.1 series
	"gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano",
	// GPT-4o series
	"gpt-4o", "gpt-4o-2024-05-13", "gpt-4o-2024-08-06", "gpt-4o-2024-11-20",
	"gpt-4o-audio-preview", "gpt-4o-realtime-preview",
	"gpt-4o-search-preview", "gpt-4o-mini-search-preview",
	"gpt-4o-mini", "gpt-4o-mini-2024-07-18", "gpt-4o-mini-audio-preview", "gpt-4o-mini-realtime-preview",
	"chatgpt-4o-latest",
	// GPT-5 series
	"gpt-5", "gpt-5-mini", "gpt-5-nano", "gpt-5-chat-latest",
	// o1 series
	"o1", "o1-2024-12-17", "o1-pro",
	"o1-preview", "o1-preview-2024-09-12",
	"o1-mini", "o1-mini-2024-09-12",
	// o3 series
	"o3", "o3-pro", "o3-deep-research", "o3-mini", "o3-mini-2025-01-31",
	// o4 series
	"o4-mini", "o4-mini-deep-research",
	// Other models
	"computer-use-preview",
	"codex-mini-latest",
	"gpt-image-1",
	// Transcription and TTS
	"gpt-4o-transcribe", "gpt-4o-mini-transcribe", "gpt-4o-mini-tts",
	// Embeddings
	"text-embedding-ada-002", "text-embedding-3-small", "text-embedding-3-large",
	// Legacy text models
	"text-curie-001", "text-babbage-001", "text-ada-001", "text-davinci-002", "text-davinci-003",
	"text-moderation-latest", "text-moderation-stable",
	"text-davinci-edit-001",
	"davinci-002", "babbage-002",
	// Image generation
	"dall-e-2", "dall-e-3",
	// Audio
	"whisper-1",
	"tts-1", "tts-1-1106", "tts-1-hd", "tts-1-hd-1106",
}
