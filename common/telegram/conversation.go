package telegram

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
)

func cancelConversationInlineKeyboard() gotgbot.InlineKeyboardMarkup {
	bt := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
			{Text: "取消", CallbackData: "cancel"},
		}},
	}

	return bt
}

func cancelConversationOpts() *handlers.ConversationOpts {
	return &handlers.ConversationOpts{
		Exits:        []ext.Handler{handlers.NewCallback(callbackquery.Equal("cancel"), cancelConversation)},
		StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		AllowReEntry: true,
	}
}

func cancelConversation(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "已取消!",
	})
	if err != nil {
		return fmt.Errorf("failed to answer start callback query: %w", err)
	}

	_, err = cb.Message.Delete(b, nil)
	if err != nil {
		return fmt.Errorf("failed to send cancel message: %w", err)
	}

	return handlers.EndConversation()
}
