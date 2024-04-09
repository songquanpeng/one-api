package notify

import (
	"context"
	"one-api/common"
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
		common.SysLog("email notifier disabled")
		return
	}
	smtp_to := viper.GetString("notify.email.smtp_to")
	emailNotifier := channel.NewEmail(smtp_to)
	AddNotifiers(emailNotifier)
	common.SysLog("email notifier enable")
}

func InitDingTalkNotifier() {
	access_token := viper.GetString("notify.dingtalk.token")
	secret := viper.GetString("notify.dingtalk.secret")
	keyWord := viper.GetString("notify.dingtalk.keyWord")
	if access_token == "" || (secret == "" && keyWord == "") {
		return
	}

	var dingTalkNotifier Notifier

	if secret != "" {
		dingTalkNotifier = channel.NewDingTalk(access_token, secret)
	} else {
		dingTalkNotifier = channel.NewDingTalkWithKeyWord(access_token, keyWord)
	}

	AddNotifiers(dingTalkNotifier)
	common.SysLog("dingtalk notifier enable")
}

func InitLarkNotifier() {
	access_token := viper.GetString("notify.lark.token")
	secret := viper.GetString("notify.lark.secret")
	keyWord := viper.GetString("notify.lark.keyWord")
	if access_token == "" || (secret == "" && keyWord == "") {
		return
	}

	var larkNotifier Notifier

	if secret != "" {
		larkNotifier = channel.NewLark(access_token, secret)
	} else {
		larkNotifier = channel.NewLarkWithKeyWord(access_token, keyWord)
	}

	AddNotifiers(larkNotifier)
	common.SysLog("lark notifier enable")
}

func InitPushdeerNotifier() {
	pushkey := viper.GetString("notify.pushdeer.pushkey")
	if pushkey == "" {
		return
	}

	pushdeerNotifier := channel.NewPushdeer(pushkey, viper.GetString("notify.pushdeer.url"))

	AddNotifiers(pushdeerNotifier)
	common.SysLog("pushdeer notifier enable")
}

func InitTelegramNotifier() {
	bot_token := viper.GetString("notify.telegram.bot_api_key")
	chat_id := viper.GetString("notify.telegram.chat_id")
	if bot_token == "" || chat_id == "" {
		return
	}

	telegramNotifier := channel.NewTelegram(bot_token, chat_id)

	AddNotifiers(telegramNotifier)
	common.SysLog("telegram notifier enable")
}
