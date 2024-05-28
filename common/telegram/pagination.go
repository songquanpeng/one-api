package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type paginationParams struct {
	key   string
	page  int
	total int
}

func paginationHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	user := getBindUser(b, ctx)
	if user == nil {
		return nil
	}

	cb := ctx.Update.CallbackQuery
	parts := strings.Split(strings.TrimPrefix(ctx.CallbackQuery.Data, "p:"), ",")
	page, err := strconv.Atoi(parts[1])
	if err != nil {
		cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "参数错误!",
		})

		return nil
	}

	switch parts[0] {
	case "apikey":
		message, pageParams := getApikeyList(user.Id, page)
		if pageParams == nil {
			cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
				Text: message,
			})
			return nil
		}

		_, _, err := cb.Message.EditText(b, message, &gotgbot.EditMessageTextOpts{
			ParseMode:   "MarkdownV2",
			ReplyMarkup: getPaginationInlineKeyboard(pageParams.key, pageParams.page, pageParams.total),
		})
		if err != nil {
			return fmt.Errorf("failed to send APIKEY message: %w", err)
		}
	default:
		cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text: "未知的类型!",
		})
	}
	return nil
}

func getPaginationInlineKeyboard(key string, page int, total int) gotgbot.InlineKeyboardMarkup {
	var bt gotgbot.InlineKeyboardMarkup
	var buttons []gotgbot.InlineKeyboardButton
	if page > 1 {
		buttons = append(buttons, gotgbot.InlineKeyboardButton{Text: fmt.Sprintf("上一页(%d/%d)", page-1, total), CallbackData: fmt.Sprintf("p:%s,%d", key, page-1)})
	}
	if page < total {
		buttons = append(buttons, gotgbot.InlineKeyboardButton{Text: fmt.Sprintf("下一页(%d/%d)", page+1, total), CallbackData: fmt.Sprintf("p:%s,%d", key, page+1)})
	}
	bt.InlineKeyboard = append(bt.InlineKeyboard, buttons)
	return bt
}

func getPageParams(key string, page, size, totalCount int) *paginationParams {
	// 根据总数计算总页数
	total := totalCount / size
	if totalCount%size > 0 {
		total++
	}

	return &paginationParams{
		page:  page,
		total: total,
		key:   key,
	}
}
