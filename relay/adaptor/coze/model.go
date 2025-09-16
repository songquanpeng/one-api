package coze

type Message struct {
	Role        string `json:"role"`
	Type        string `json:"type,omitempty"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
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

type V3StreamResponse struct {
	Id             string `json:"id"`
	ConversationId string `json:"conversation_id"`
	BotId          string `json:"bot_id"`
	Role           string `json:"role"`
	Type           string `json:"type"`
	Content        string `json:"content"`
	ContentType    string `json:"content_type"`
	ChatId         string `json:"chat_id"`
	CreatedAt      int    `json:"created_at"`
	CompletedAt    int    `json:"completed_at"`
	LastError      struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"last_error"`
	Status string `json:"status"`
	Usage  struct {
		TokenCount  int `json:"token_count"`
		OutputCount int `json:"output_count"`
		InputCount  int `json:"input_count"`
	} `json:"usage"`
	SectionId string `json:"section_id"`
}

type V3Response struct {
	Data struct {
		Id             string `json:"id"`
		ConversationId string `json:"conversation_id"`
		BotId          string `json:"bot_id"`
		Content        string `json:"content"`
		ContentType    string `json:"content_type"`
		CreatedAt      int    `json:"created_at"`
		LastError      struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"last_error"`
		Status string `json:"status"`
	} `json:"data"`
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type V3Request struct {
	BotId              string    `json:"bot_id"`
	UserId             string    `json:"user_id"`
	AdditionalMessages []Message `json:"additional_messages"`
	Stream             bool      `json:"stream"`
}
