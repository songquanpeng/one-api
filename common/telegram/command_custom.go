package telegram

import (
	"fmt"
	"html"
	"one-api/model"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func commandCustom(b *gotgbot.Bot, ctx *ext.Context) error {
	command := strings.TrimSpace(ctx.EffectiveMessage.Text)
	// 去除/
	command = strings.TrimPrefix(command, "/")

	menu, err := model.GetTelegramMenuByCommand(command)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "系统错误，请稍后再试", nil)
		return nil
	}

	if menu == nil {
		ctx.EffectiveMessage.Reply(b, "未找到该命令", nil)
		return nil
	}

	_, err = b.SendMessage(ctx.EffectiveSender.Id(), menu.ReplyMessage, &gotgbot.SendMessageOpts{
		ParseMode: menu.ParseMode,
	})

	if err != nil {
		return fmt.Errorf("failed to send %s message: %w", command, err)
	}

	return nil
}

func escapeText(text, parseMode string) string {
	switch parseMode {
	case "MarkdownV2":
		// Characters that need to be escaped in MarkdownV2 mode
		chars := []string{"_", "*", "[", "]", "(", ")", "~", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		for _, char := range chars {
			text = strings.ReplaceAll(text, char, "\\"+char)
		}
	case "HTML":
		// Escape HTML special characters
		text = html.EscapeString(text)
		// Markdown mode does not require escaping
	}
	return text
}
