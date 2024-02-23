package telegram

import (
	"one-api/model"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func commandUnbindStart(b *gotgbot.Bot, ctx *ext.Context) error {
	user := getBindUser(b, ctx)
	if user == nil {
		return nil
	}

	updateUser := map[string]interface{}{
		"telegram_id": 0,
	}

	err := model.UpdateUser(user.Id, updateUser)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "绑定失败，请稍后再试", nil)
		return handlers.EndConversation()
	}

	ctx.EffectiveMessage.Reply(b, "解邦成功", nil)
	return nil
}
