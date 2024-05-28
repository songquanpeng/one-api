package telegram

import (
	"fmt"
	"net/url"
	"one-api/common/config"
	"one-api/model"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func commandApikeyStart(b *gotgbot.Bot, ctx *ext.Context) error {
	user := getBindUser(b, ctx)
	if user == nil {
		return nil
	}

	message, pageParams := getApikeyList(user.Id, 1)
	if pageParams == nil {
		_, err := ctx.EffectiveMessage.Reply(b, message, nil)
		if err != nil {
			return fmt.Errorf("failed to send APIKEY message: %w", err)
		}
		return nil
	}

	_, err := ctx.EffectiveMessage.Reply(b, message, &gotgbot.SendMessageOpts{
		ParseMode:   "MarkdownV2",
		ReplyMarkup: getPaginationInlineKeyboard(pageParams.key, pageParams.page, pageParams.total),
	})
	if err != nil {
		return fmt.Errorf("failed to send APIKEY message: %w", err)
	}

	return nil
}

func getApikeyList(userId, page int) (message string, pageParams *paginationParams) {
	genericParams := &model.GenericParams{
		PaginationParams: model.PaginationParams{
			Page: page,
			Size: 5,
		},
	}

	list, err := model.GetUserTokensList(userId, genericParams)

	if err != nil {
		return "系统错误，请稍后再试", nil
	}

	if list.Data == nil || len(*list.Data) == 0 {
		return "找不到令牌", nil
	}

	chatUrlTmp := ""
	if config.ServerAddress != "" {
		chatUrlTmp = getChatUrl()
	}

	message = "点击令牌可复制：\n"

	for _, token := range *list.Data {
		key := "sk-" + token.Key
		message += fmt.Sprintf("*%s* : `%s`\n", escapeText(token.Name, "MarkdownV2"), key)
		if chatUrlTmp != "" {
			message += strings.ReplaceAll(chatUrlTmp, `setToken`, key)
		}
		message += "\n"
	}

	return message, getPageParams("apikey", page, genericParams.Size, int(list.TotalCount))
}

func getChatUrl() string {
	serverAddress := strings.TrimSuffix(config.ServerAddress, "/")
	chatNextUrl := fmt.Sprintf(`{"key":"setToken","url":"%s"}`, serverAddress)
	chatNextUrl = "https://chat.oneapi.pro/#/?settings=" + url.QueryEscape(chatNextUrl)
	if config.ChatLink != "" {
		chatLink := strings.TrimSuffix(config.ChatLink, "/")
		chatNextUrl = strings.ReplaceAll(chatNextUrl, `https://chat.oneapi.pro`, chatLink)
	}

	jumpUrl := fmt.Sprintf(`%s/jump?url=`, serverAddress)

	amaUrl := jumpUrl + url.QueryEscape(fmt.Sprintf(`ama://set-api-key?server=%s&key=setToken`, serverAddress))

	openCatUrl := jumpUrl + url.QueryEscape(fmt.Sprintf(`opencat://team/join?domain=%s&token=setToken`, serverAddress))

	return fmt.Sprintf("[Next Chat](%s)  [AMA](%s)  [OpenCat](%s)\n", chatNextUrl, amaUrl, openCatUrl)
}
