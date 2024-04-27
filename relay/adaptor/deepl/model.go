package deepl

type Request struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
}

type Translation struct {
	DetectedSourceLanguage string `json:"detected_source_language,omitempty"`
	Text                   string `json:"text,omitempty"`
}

type Response struct {
	Translations []Translation `json:"translations,omitempty"`
	Message      string        `json:"message,omitempty"`
}
