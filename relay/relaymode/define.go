package relaymode

const (
	Unknown = iota
	ChatCompletions
	Completions
	Embeddings
	Moderations
	ImagesGenerations
	Edits
	AudioSpeech
	AudioTranscription
	AudioTranslation
	// Responses is the OpenAI Responses API (/v1/responses)
	Responses
	// Proxy is a special relay mode for proxying requests to custom upstream
	Proxy
)
