package notify

import (
	"context"
	"one-api/common/logger"
	"one-api/common/notify/channel"

	"github.com/spf13/viper"
)

type Notifier interface {
	Send(context.Context, string, string) error
	Name() string
}

func InitNotifier() {
	InitEmailNotifier()
	InitDingTalkNotifier()
	InitLarkNotifier()
	InitPushdeerNotifier()
	InitTelegramNotifier()
}

func InitEmailNotifier() {
	if viper.GetBool("notify.email.disable") {
		logger.SysLog("email notifier disabled")
		return
	}
	smtpTo := viper.GetString("notify.email.smtp_to")
	emailNotifier := channel.NewEmail(smtpTo)
	AddNotifiers(emailNotifier)
	logger.SysLog("email notifier enable")
}

func InitDingTalkNotifier() {
	accessToken := viper.GetString("notify.dingtalk.token")
	secret := viper.GetString("notify.dingtalk.secret")
	keyWord := viper.GetString("notify.dingtalk.keyWord")
	if accessToken == "" || (secret == "" && keyWord == "") {
		return
	}

	var dingTalkNotifier Notifier

	if secret != "" {
		dingTalkNotifier = channel.NewDingTalk(accessToken, secret)
	} else {
		dingTalkNotifier = channel.NewDingTalkWithKeyWord(accessToken, keyWord)
	}

	AddNotifiers(dingTalkNotifier)
	logger.SysLog("dingtalk notifier enable")
}

func InitLarkNotifier() {
	accessToken := viper.GetString("notify.lark.token")
	secret := viper.GetString("notify.lark.secret")
	keyWord := viper.GetString("notify.lark.keyWord")
	if accessToken == "" || (secret == "" && keyWord == "") {
		return
	}

	var larkNotifier Notifier

	if secret != "" {
		larkNotifier = channel.NewLark(accessToken, secret)
	} else {
		larkNotifier = channel.NewLarkWithKeyWord(accessToken, keyWord)
	}

	AddNotifiers(larkNotifier)
	logger.SysLog("lark notifier enable")
}

func InitPushdeerNotifier() {
	pushkey := viper.GetString("notify.pushdeer.pushkey")
	if pushkey == "" {
		return
	}

	pushdeerNotifier := channel.NewPushdeer(pushkey, viper.GetString("notify.pushdeer.url"))

	AddNotifiers(pushdeerNotifier)
	logger.SysLog("pushdeer notifier enable")
}

func InitTelegramNotifier() {
	botToken := viper.GetString("notify.telegram.bot_api_key")
	chatId := viper.GetString("notify.telegram.chat_id")
	httpProxy := viper.GetString("notify.telegram.http_proxy")
	if botToken == "" || chatId == "" {
		return
	}

	telegramNotifier := channel.NewTelegram(botToken, chatId, httpProxy)

	AddNotifiers(telegramNotifier)
	logger.SysLog("telegram notifier enable")
}
