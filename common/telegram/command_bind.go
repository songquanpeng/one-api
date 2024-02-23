package telegram

import (
	"fmt"
	"one-api/model"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func commandBindInit() (handler ext.Handler) {
	return handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("bind", commandBindStart)},
		map[string][]ext.Handler{
			"token": {handlers.NewMessage(noCommands, commandBindToken)},
		},
		cancelConversationOpts(),
	)
}

func commandBindStart(b *gotgbot.Bot, ctx *ext.Context) error {
	user := getBindUser(b, ctx)
	if user != nil {
		ctx.EffectiveMessage.Reply(b, "您的账户已绑定，请解邦后再试", nil)
		return handlers.EndConversation()
	}

	_, err := ctx.EffectiveMessage.Reply(b, "请输入你的访问令牌", &gotgbot.SendMessageOpts{
		ParseMode:   "html",
		ReplyMarkup: cancelConversationInlineKeyboard(),
	})
	if err != nil {
		return fmt.Errorf("failed to send bind start message: %w", err)
	}
	return handlers.NextConversationState("token")

}

func commandBindToken(b *gotgbot.Bot, ctx *ext.Context) error {
	tgUserId := getTGUserId(b, ctx)
	if tgUserId == 0 {
		return handlers.EndConversation()
	}

	input := ctx.EffectiveMessage.Text
	// 去除input前后空格
	input = strings.TrimSpace(input)

	user := model.ValidateAccessToken(input)
	if user == nil {
		// If the number is not valid, try again!
		ctx.EffectiveMessage.Reply(b, "Token 错误，请重试", &gotgbot.SendMessageOpts{
			ParseMode:   "html",
			ReplyMarkup: cancelConversationInlineKeyboard(),
		})
		// We try the age handler again
		return handlers.NextConversationState("token")
	}

	if user.TelegramId != 0 {
		ctx.EffectiveMessage.Reply(b, "您的账户已绑定，请解邦后再试", nil)
		return handlers.EndConversation()
	}

	// 查询该tg用户是否已经绑定其他账户
	if model.IsTelegramIdAlreadyTaken(tgUserId) {
		ctx.EffectiveMessage.Reply(b, "该TG已绑定其他账户，请解邦后再试", nil)
		return handlers.EndConversation()
	}

	// 绑定
	updateUser := model.User{
		Id:         user.Id,
		TelegramId: tgUserId,
	}
	err := updateUser.Update(false)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "绑定失败，请稍后再试", nil)
		return handlers.EndConversation()
	}

	_, err = ctx.EffectiveMessage.Reply(b, "绑定成功", nil)
	if err != nil {
		return fmt.Errorf("failed to send bind token message: %w", err)
	}
	return handlers.EndConversation()
}
