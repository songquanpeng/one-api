package coze

import "one-api/types"

type CozeStatus struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type CozeRequest struct {
	BotID          string        `json:"bot_id"`
	Query          string        `json:"query"`
	Stream         bool          `json:"stream"`
	User           string        `json:"user"`
	ConversationID string        `json:"conversation_id"`
	ChatHistory    []CozeMessage `json:"chat_history"`
}

type CozeMessage struct {
	Role        string `json:"role"`
	Type        string `json:"type,omitempty"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type CozeResponse struct {
	CozeStatus
	ConversationID string        `json:"conversation_id"`
	Messages       []CozeMessage `json:"messages"`
}

func (cr *CozeResponse) String() string {
	message := ""

	for _, msg := range cr.Messages {
		if msg.Type == "answer" && msg.Role == types.ChatMessageRoleAssistant {
			message = msg.Content
			break
		}
	}

	return message
}

type CozeStreamResponse struct {
	Event            string      `json:"event"`
	ErrorInformation string      `json:"error_information,omitempty"`
	Message          CozeMessage `json:"message,omitempty"`
	IsFinish         bool        `json:"is_finish,omitempty"`
	Index            int         `json:"index,omitempty"`
	ConversationID   string      `json:"conversation_id,omitempty"`
}
