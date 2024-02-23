package telegram

import (
	"fmt"
	"one-api/model"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func commandRechargeInit() (handler ext.Handler) {
	return handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("recharge", commandRechargeStart)},
		map[string][]ext.Handler{
			"recharge_token": {handlers.NewMessage(noCommands, commandRechargeToken)},
		},
		cancelConversationOpts(),
	)
}

func commandRechargeStart(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "请输入你的兑换码", &gotgbot.SendMessageOpts{
		ParseMode:   "html",
		ReplyMarkup: cancelConversationInlineKeyboard(),
	})
	if err != nil {
		return fmt.Errorf("failed to send recharge start message: %w", err)
	}
	return handlers.NextConversationState("recharge_token")

}

func commandRechargeToken(b *gotgbot.Bot, ctx *ext.Context) error {
	user := getBindUser(b, ctx)
	if user == nil {
		return handlers.EndConversation()
	}

	input := ctx.EffectiveMessage.Text
	// 去除input前后空格
	input = strings.TrimSpace(input)

	quota, err := model.Redeem(input, user.Id)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "充值失败："+err.Error(), nil)
		return handlers.EndConversation()
	}

	money := fmt.Sprintf("%.2f", float64(quota)/500000)
	_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("成功充值 $%s ", money), nil)
	if err != nil {
		return fmt.Errorf("failed to send recharge token message: %w", err)
	}
	return handlers.EndConversation()
}
