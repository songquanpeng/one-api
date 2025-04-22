package main

import (
	"fmt"
	"github.com/songquanpeng/one-api/model"
	onesdk "github.com/songquanpeng/one-api/sdk/api"
	"log"
	"strconv"
)

func main() {
	// for test
	config := onesdk.Config{
		Host: "http://127.0.0.1",
		Port: 3000,
		Key:  "123456789012345678901234567890",
	}
	client := onesdk.OneClient{
		Config: &config,
	}

	// 用户API使用测试
	// 添加用户
	user := onesdk.User{
		Username:    "user1",
		DisplayName: "user1",
		Password:    "user1@123_%6",
	}
	err := user.Add(&client)
	if err != nil {
		log.Fatal("add user err=", err)
	}
	fmt.Println("add user:", user, "success！")
	// 查找用户
	users := onesdk.Users{}
	// 可根据用户名、显示名、邮箱、手机号等信息进行模糊查询
	users.Query = map[string]string{
		"keyword": "user1",
	}
	err = users.List(&client)
	if err != nil {
		log.Fatal("list user err=", err)
	}
	tmpUser := onesdk.User{}
	for i, u := range users.Users {
		// 删除的不显示
		if u.Status == model.UserStatusDeleted {
			continue
		}
		fmt.Println("user["+strconv.Itoa(i)+"]:", *u)
		if u.Username == "user1" {
			tmpUser = *u
		}
	}
	fmt.Println("list user success！")
	// 获取用户
	user = onesdk.User{}
	user.ID = tmpUser.ID
	err = user.Get(&client)
	if err != nil {
		log.Fatal("get user err=", err)
	}
	fmt.Println("get user:", user, "success！")
	//更新用户
	user.Quota = 500000000
	err = user.Update(&client)
	if err != nil {
		log.Fatal("update user err=", err)
	}
	fmt.Println("update user:", user, "success！\r\n")

	// 渠道API使用测试
	channel := onesdk.Channel{
		Name: "ch1",
		ChannelConfig: onesdk.ChannelConfig{
			Region: "",
			Sk:     "",
			Ak:     "",
		},
		Group:        "default",
		Models:       "moonshot-v1-8k,moonshot-v1-32k,moonshot-v1-128k",
		ModelMapping: "",
		Other:        "",
		SystemPrompt: "",
		Type:         25,
		Key:          "key",
	}
	err = channel.Add(&client)
	if err != nil {
		log.Fatal("add channel err=", err)
	}
	fmt.Println("add channel:", channel, "success！")
	// 查询渠道
	channels := onesdk.Channels{}
	err = channels.List(&client)
	channels.Query = map[string]string{
		"keyword": "ch1",
	}
	if err != nil {
		log.Fatal("list channel err=", err)
	}
	tmpChannel := onesdk.Channel{}
	for i, c := range channels.Channels {
		fmt.Println("channel["+strconv.Itoa(i)+"]:", *c)
		if c.Name == "ch1" {
			tmpChannel = *c
		}
	}
	fmt.Println("list channel success！")
	// 更新渠道
	updateChannel := tmpChannel
	updateChannel.Name = "ch1-updated"
	err = updateChannel.Update(&client)
	if err != nil {
		log.Fatal("update channel err=", err)
	}
	fmt.Println("update channel:", updateChannel, "success！")
	// 获取渠道
	channel = onesdk.Channel{}
	channel.ID = tmpChannel.ID
	err = channel.Get(&client)
	if err != nil {
		log.Fatal("get channel err=", err)
	}
	fmt.Println("get channel:", channel, "success！")
	// 测试渠道（模型）是否正常
	err = channel.Test(&client)
	if err != nil {
		log.Fatal("test channel err=", err)
	}
	fmt.Println("test channel:", channel, "success！")
	// 删除渠道
	err = updateChannel.Delete(&client)
	if err != nil {
		log.Fatal("delete channel err=", err)
	}
	fmt.Println("delete channel:", updateChannel, "success！\r\n")

	// 令牌API使用测试
	// 添加令牌
	token := onesdk.Token{
		Name:           "token1",
		UserID:         user.ID,
		Models:         "/data/DeepSeek-R1,ERNIE-3.5-8K",
		RemainQuota:    5000000000,
		UnlimitedQuota: false,
		ExpiredTime:    -1,
		Subnet:         "",
	}
	err = token.Add(&client)
	if err != nil {
		log.Fatal("add token err=", err)
	}
	fmt.Println("add token:", token, "success！")
	//查询令牌
	tokens := onesdk.Tokens{}
	tokens.UserID = user.ID
	tokens.Query = map[string]string{
		"keyword": "token1",
	}
	err = tokens.List(&client)
	if err != nil {
		log.Fatal("list token err=", err)
	}
	tmpToken := onesdk.Token{}
	for i, t := range tokens.Tokens {
		fmt.Println("token["+strconv.Itoa(i)+"]:", *t)
		if t.Name == "token1" {
			tmpToken = *t
		}
	}
	//更新令牌
	token = tmpToken
	token.Models = "/data/DeepSeek-R1"
	token.RemainQuota = 9009000000
	err = token.Update(&client)
	if err != nil {
		log.Fatal("update token err=", err)
	}
	fmt.Println("update token:", token, "success！")
	// 获取token
	token = onesdk.Token{ID: token.ID, UserID: tmpToken.UserID}
	err = token.Get(&client)
	if err != nil {
		log.Fatal("get token err=", err)
	}
	fmt.Println("get token:", token, "success！")
	// delete token
	err = token.Delete(&client)
	if err != nil {
		log.Fatal("delete token err=", err)
	}
	fmt.Println("delete token:", token, "success！\r\n")

	// 日志API使用测试
	logs := onesdk.Logs{}
	logs.Query = map[string]string{
		"username": "user1",
	}
	err = logs.Get(&client)
	if err != nil {
		log.Fatal(err)
	}
	for i, l := range logs.Logs {
		fmt.Println("log["+strconv.Itoa(i)+"]=", *l)
	}
	fmt.Println("get logs success！\r\n\r\n")

	// 删除用户
	err = user.Delete(&client)
	if err != nil {
		log.Fatal("delete user err=", err)
	}
	fmt.Println("delete user:", user, "success！")

	// 操作root自己的令牌
	rootToken := onesdk.Token{
		Name:           "token1",
		Models:         "/data/DeepSeek-R1,ERNIE-3.5-8K",
		RemainQuota:    5000000000,
		UnlimitedQuota: false,
		ExpiredTime:    -1,
		Subnet:         "",
	}
	err = rootToken.Add(&client)
	if err != nil {
		log.Fatal("add root token err=", err)
	}
	fmt.Println("add root token:", rootToken, "success！")
	//查询令牌
	tokens = onesdk.Tokens{}
	tokens.Query = map[string]string{
		"keyword": "token1",
	}
	err = tokens.List(&client)
	if err != nil {
		log.Fatal("list root token err=", err)
	}
	tmpToken = onesdk.Token{}
	for i, t := range tokens.Tokens {
		fmt.Println("token["+strconv.Itoa(i)+"]:", *t)
		if t.Name == "token1" {
			tmpToken = *t
		}
	}
	// 获取令牌
	rootToken = onesdk.Token{ID: tmpToken.ID}
	err = rootToken.Get(&client)
	if err != nil {
		log.Fatal("get root token err=", err)
	}
	fmt.Println("get root token:", rootToken, "success！")
	// 删除令牌
	err = rootToken.Delete(&client)
	if err != nil {
		log.Fatal("delete root token err=", err)
	}
	fmt.Println("delete root token:", rootToken, "success！")
}
