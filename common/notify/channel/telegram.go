package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common/requester"
	"one-api/types"
)

const telegramURL = "https://api.telegram.org/bot"

type Telegram struct {
	secret string
	chatID string
}

type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type telegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

func NewTelegram(secret string, chatID string) *Telegram {
	return &Telegram{
		secret: secret,
		chatID: chatID,
	}
}

func (t *Telegram) Name() string {
	return "Telegram"
}

func (t *Telegram) Send(ctx context.Context, title, message string) error {
	const maxMessageLength = 4096
	message = fmt.Sprintf("*%s*\n%s", title, message)
	messages := splitTelegramMessageIntoParts(message, maxMessageLength)

	client := requester.NewHTTPRequester("", telegramErrFunc)
	client.Context = ctx
	client.IsOpenAI = false

	for _, msg := range messages {
		err := t.sendMessage(msg, client)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Telegram) sendMessage(message string, client *requester.HTTPRequester) error {
	msg := telegramMessage{
		ChatID:    t.chatID,
		Text:      message,
		ParseMode: "Markdown",
	}

	uri := telegramURL + t.secret + "/sendMessage"

	req, err := client.NewRequest(http.MethodPost, uri, client.WithHeader(requester.GetJsonHeaders()), client.WithBody(msg))
	if err != nil {
		return err
	}

	resp, errWithOP := client.SendRequestRaw(req)
	if errWithOP != nil {
		return fmt.Errorf("%s", errWithOP.Message)
	}
	defer resp.Body.Close()

	telegramErr := telegramErrFunc(resp)
	if telegramErr != nil {
		return fmt.Errorf("%s", telegramErr.Message)
	}

	return nil
}

func splitTelegramMessageIntoParts(message string, partSize int) []string {
	var parts []string
	for len(message) > partSize {
		parts = append(parts, message[:partSize])
		message = message[partSize:]
	}
	parts = append(parts, message)

	return parts
}

func telegramErrFunc(resp *http.Response) *types.OpenAIError {
	respMsg := &telegramResponse{}
	err := json.NewDecoder(resp.Body).Decode(respMsg)
	if err != nil {
		return nil
	}

	if respMsg.Ok {
		return nil
	}

	return &types.OpenAIError{
		Message: fmt.Sprintf("send msg err. err msg: %s", respMsg.Description),
		Type:    "telegram_error",
	}
}
