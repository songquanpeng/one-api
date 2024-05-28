package telegram

import (
	"one-api/common/config"
	"one-api/common/utils"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func commandAffStart(b *gotgbot.Bot, ctx *ext.Context) error {
	user := getBindUser(b, ctx)
	if user == nil {
		return nil
	}

	if user.AffCode == "" {
		user.AffCode = utils.GetRandomString(4)
		if err := user.Update(false); err != nil {
			ctx.EffectiveMessage.Reply(b, "系统错误，请稍后再试", nil)
			return nil
		}
	}

	messae := "您可以通过分享您的邀请码来邀请朋友，每次成功邀请将获得奖励。\n\n您的邀请码是: " + user.AffCode
	if config.ServerAddress != "" {
		serverAddress := strings.TrimSuffix(config.ServerAddress, "/")
		messae += "\n\n页面地址：" + serverAddress + "/register?aff=" + user.AffCode
	}

	ctx.EffectiveMessage.Reply(b, messae, nil)

	return nil
}
