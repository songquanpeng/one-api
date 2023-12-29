package baichuan

import "one-api/providers/openai"

type BaichuanMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaichuanKnowledgeBase struct {
	Ids []string `json:"id"`
}

type BaichuanChatRequest struct {
	Model             string                `json:"model"`
	Messages          []BaichuanMessage     `json:"messages"`
	Stream            bool                  `json:"stream,omitempty"`
	Temperature       float64               `json:"temperature,omitempty"`
	TopP              float64               `json:"top_p,omitempty"`
	TopK              int                   `json:"top_k,omitempty"`
	WithSearchEnhance bool                  `json:"with_search_enhance,omitempty"`
	KnowledgeBase     BaichuanKnowledgeBase `json:"knowledge_base,omitempty"`
}

type BaichuanKnowledgeBaseResponse struct {
	Cites []struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		FileId  string `json:"file_id"`
	} `json:"cites"`
}

type BaichuanChatResponse struct {
	openai.OpenAIProviderChatResponse
	KnowledgeBase BaichuanKnowledgeBaseResponse `json:"knowledge_base,omitempty"`
}
