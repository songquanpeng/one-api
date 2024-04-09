package channel_test

import (
	"context"
	"fmt"
	"testing"

	"one-api/common/notify/channel"
	"one-api/common/requester"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func InitConfig() {
	viper.AddConfigPath("/one-api")
	viper.SetConfigName("config")
	viper.ReadInConfig()
	requester.InitHttpClient()
}

func TestDingTalkSend(t *testing.T) {
	InitConfig()
	access_token := viper.GetString("notify.dingtalk.token")
	secret := viper.GetString("notify.dingtalk.secret")
	dingTalk := channel.NewDingTalk(access_token, secret)

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Nil(t, err)
}

func TestDingTalkSendWithKeyWord(t *testing.T) {
	InitConfig()
	access_token := viper.GetString("notify.dingtalk.token")
	keyWord := viper.GetString("notify.dingtalk.keyWord")

	dingTalk := channel.NewDingTalkWithKeyWord(access_token, keyWord)

	err := dingTalk.Send(context.Background(), "Test Title", "Test Message")
	assert.Nil(t, err)
}

func TestDingTalkSendError(t *testing.T) {
	InitConfig()
	access_token := viper.GetString("notify.dingtalk.token")
	secret := "test"
	dingTalk := channel.NewDingTalk(access_token, secret)

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Error(t, err)
}

func TestLarkSend(t *testing.T) {
	InitConfig()
	access_token := viper.GetString("notify.lark.token")
	secret := viper.GetString("notify.lark.secret")
	dingTalk := channel.NewLark(access_token, secret)

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Nil(t, err)
}

func TestLarkSendWithKeyWord(t *testing.T) {
	InitConfig()
	access_token := viper.GetString("notify.lark.token")
	keyWord := viper.GetString("notify.lark.keyWord")

	dingTalk := channel.NewLarkWithKeyWord(access_token, keyWord)

	err := dingTalk.Send(context.Background(), "Test Title", "Test Message\n\n- 111\n- 222")
	assert.Nil(t, err)
}

func TestLarkSendError(t *testing.T) {
	InitConfig()
	access_token := viper.GetString("notify.lark.token")
	secret := "test"
	dingTalk := channel.NewLark(access_token, secret)

	err := dingTalk.Send(context.Background(), "Title", "*Message*")
	fmt.Println(err)
	assert.Error(t, err)
}

func TestPushdeerSend(t *testing.T) {
	InitConfig()
	pushkey := viper.GetString("notify.pushdeer.pushkey")
	dingTalk := channel.NewPushdeer(pushkey, "")

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Nil(t, err)
}

func TestPushdeerSendError(t *testing.T) {
	InitConfig()
	pushkey := "test"
	dingTalk := channel.NewPushdeer(pushkey, "")

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Error(t, err)
}

func TestTelegramSend(t *testing.T) {
	InitConfig()
	secret := viper.GetString("notify.telegram.bot_api_key")
	chatID := viper.GetString("notify.telegram.chat_id")
	dingTalk := channel.NewTelegram(secret, chatID)

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Nil(t, err)
}

func TestTelegramSendError(t *testing.T) {
	InitConfig()
	secret := "test"
	chatID := viper.GetString("notify.telegram.chat_id")
	dingTalk := channel.NewTelegram(secret, chatID)

	err := dingTalk.Send(context.Background(), "Test Title", "*Test Message*")
	fmt.Println(err)
	assert.Error(t, err)
}
