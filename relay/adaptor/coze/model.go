package coze

type Message struct {
	Role        string `json:"role"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type ErrorInformation struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Request struct {
	ConversationId string    `json:"conversation_id,omitempty"`
	BotId          string    `json:"bot_id"`
	User           string    `json:"user"`
	Query          string    `json:"query"`
	ChatHistory    []Message `json:"chat_history,omitempty"`
	Stream         bool      `json:"stream"`
}

type Response struct {
	ConversationId string    `json:"conversation_id,omitempty"`
	Messages       []Message `json:"messages,omitempty"`
	Code           int       `json:"code,omitempty"`
	Msg            string    `json:"msg,omitempty"`
}

type StreamResponse struct {
	Event            string            `json:"event,omitempty"`
	Message          *Message          `json:"message,omitempty"`
	IsFinish         bool              `json:"is_finish,omitempty"`
	Index            int               `json:"index,omitempty"`
	ConversationId   string            `json:"conversation_id,omitempty"`
	ErrorInformation *ErrorInformation `json:"error_information,omitempty"`
}
