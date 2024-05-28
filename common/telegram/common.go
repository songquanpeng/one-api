package telegram

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/model"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"
)

var TGupdater *ext.Updater
var TGBot *gotgbot.Bot
var TGDispatcher *ext.Dispatcher
var TGWebHookSecret = ""
var TGEnabled = false

func InitTelegramBot() {
	if TGEnabled {
		logger.SysLog("Telegram bot has been started")
		return
	}

	botKey := viper.GetString("tg.bot_api_key")
	if botKey == "" {
		logger.SysLog("Telegram bot is not enabled")
		return
	}

	var err error
	TGBot, err = gotgbot.NewBot(botKey, getBotOpts())
	if err != nil {
		logger.SysLog("failed to create new telegram bot: " + err.Error())
		return
	}

	TGDispatcher = setDispatcher()
	TGupdater = ext.NewUpdater(TGDispatcher, nil)

	StartTelegramBot()
}

func StartTelegramBot() {
	botWebhook := viper.GetString("tg.webhook_secret")
	if botWebhook != "" {
		if config.ServerAddress == "" {
			logger.SysLog("Telegram bot is not enabled: Server address is not set")
			StopTelegramBot()
			return
		}
		TGWebHookSecret = botWebhook
		serverAddress := strings.TrimSuffix(config.ServerAddress, "/")
		urlPath := fmt.Sprintf("/api/telegram/%s", viper.GetString("tg.bot_api_key"))

		webHookOpts := &ext.AddWebhookOpts{
			SecretToken: TGWebHookSecret,
		}

		err := TGupdater.AddWebhook(TGBot, urlPath, webHookOpts)
		if err != nil {
			logger.SysLog("Telegram bot failed to add webhook:" + err.Error())
			return
		}

		err = TGupdater.SetAllBotWebhooks(serverAddress, &gotgbot.SetWebhookOpts{
			MaxConnections:     100,
			DropPendingUpdates: true,
			SecretToken:        TGWebHookSecret,
		})
		if err != nil {
			logger.SysLog("Telegram bot failed to set webhook:" + err.Error())
			return
		}
	} else {
		err := TGupdater.StartPolling(TGBot, &ext.PollingOpts{
			EnableWebhookDeletion: true,
			DropPendingUpdates:    true,
			GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
				Timeout: 9,
				RequestOpts: &gotgbot.RequestOpts{
					Timeout: time.Second * 10,
				},
			},
		})

		if err != nil {
			logger.SysLog("Telegram bot failed to start polling:" + err.Error())
		}
	}

	// Idle, to keep updates coming in, and avoid bot stopping.
	go TGupdater.Idle()
	logger.SysLog(fmt.Sprintf("Telegram bot %s has been started...:", TGBot.User.Username))
	TGEnabled = true
}

func ReloadMenuAndCommands() error {
	if !TGEnabled || TGupdater == nil {
		return errors.New("telegram bot is not enabled")
	}

	menus := getMenu()
	TGBot.SetMyCommands(menus, nil)
	TGDispatcher.RemoveGroup(0)
	initCommand(TGDispatcher, menus)

	return nil
}

func StopTelegramBot() {
	if TGEnabled {
		TGupdater.Stop()
		TGupdater = nil
		TGEnabled = false
	}
}

func setDispatcher() *ext.Dispatcher {
	menus := getMenu()
	TGBot.SetMyCommands(menus, nil)

	// Create dispatcher.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			logger.SysLog("telegram an error occurred while handling update: " + err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	initCommand(dispatcher, menus)

	return dispatcher
}

func initCommand(dispatcher *ext.Dispatcher, menu []gotgbot.BotCommand) {
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("p:"), paginationHandler))
	for _, command := range menu {
		switch command.Command {
		case "bind":
			dispatcher.AddHandler(commandBindInit())
		case "unbind":
			dispatcher.AddHandler(handlers.NewCommand("unbind", commandUnbindStart))
		case "balance":
			dispatcher.AddHandler(handlers.NewCommand("balance", commandBalanceStart))
		case "recharge":
			dispatcher.AddHandler(commandRechargeInit())
		case "apikey":
			dispatcher.AddHandler(handlers.NewCommand("apikey", commandApikeyStart))
		case "aff":
			dispatcher.AddHandler(handlers.NewCommand("aff", commandAffStart))
		default:
			dispatcher.AddHandler(handlers.NewCommand(command.Command, commandCustom))
		}
	}
}

func getMenu() []gotgbot.BotCommand {
	defaultMenu := GetDefaultMenu()
	customMenu, err := model.GetTelegramMenus()

	if err != nil {
		logger.SysLog("Failed to get custom menu, error: " + err.Error())
	}

	if len(customMenu) > 0 {
		// 追加自定义菜单
		for _, menu := range customMenu {
			defaultMenu = append(defaultMenu, gotgbot.BotCommand{Command: menu.Command, Description: menu.Description})
		}
	}

	return defaultMenu
}

// 菜单 1. 绑定 2. 解绑 3. 查询余额 4. 充值 5. 获取API_KEY
func GetDefaultMenu() []gotgbot.BotCommand {
	return []gotgbot.BotCommand{
		{Command: "bind", Description: "绑定账号"},
		{Command: "unbind", Description: "解绑账号"},
		{Command: "balance", Description: "查询余额"},
		{Command: "recharge", Description: "充值"},
		{Command: "apikey", Description: "获取API_KEY"},
		{Command: "aff", Description: "获取邀请链接"},
	}
}

func noCommands(msg *gotgbot.Message) bool {
	return message.Text(msg) && !message.Command(msg)
}

func getTGUserId(b *gotgbot.Bot, ctx *ext.Context) int64 {
	if ctx.EffectiveSender.User == nil {
		ctx.EffectiveMessage.Reply(b, "无法使用命令", nil)
		return 0
	}

	return ctx.EffectiveSender.User.Id
}

func getBindUser(b *gotgbot.Bot, ctx *ext.Context) *model.User {
	tgUserId := getTGUserId(b, ctx)
	if tgUserId == 0 {
		return nil
	}

	user, err := model.GetUserByTelegramId(tgUserId)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "您的账户未绑定", nil)
		return nil
	}

	return user
}

func getHttpClient() (httpClient *http.Client) {
	proxyAddr := viper.GetString("tg.http_proxy") // http/socks5
	if proxyAddr == "" {
		return
	}

	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		logger.SysLog("failed to parse TG proxy URL: " + err.Error())
		return
	}
	switch proxyURL.Scheme {
	case "http", "https":
		httpClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	case "socks5":
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			logger.SysLog("failed to create TG SOCKS5 dialer: " + err.Error())
			return
		}
		httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.(proxy.ContextDialer).DialContext(ctx, network, addr)
				},
			},
		}
	default:
		logger.SysLog("unknown TG proxy type: " + proxyAddr)
	}

	return
}

func getBotOpts() *gotgbot.BotOpts {
	httpClient := getHttpClient()
	if httpClient == nil {
		return nil
	}

	return &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			Client: *httpClient,
		},
	}
}
