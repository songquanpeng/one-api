## golang sdk 使用方式
### 1. 安装

```
import sdk oneapi "github.com/songquanpeng/one-api/sdk/api"
``` 
### 2. 初始化

```
config := oneapi.Config{
		Host: "http://127.0.0.1",     //实际使用时请替换为实际的host
		Port: 3000,                   //实际使用时请替换为实际的port
		Key:  "12345678901234567890", //实际使用时请替换为root用户下生成的系统访问令牌
	}
	client := oneapi.OneClient{
		Config: &config,
}
```
### 3. 调用

```
(1) 用户操作
    // 添加用户
	user := oneapi.User{
		Username:    "test1",
		DisplayName: "test1",
		Password:    "test@123_%6",
	}
	err := user.Add(&client)
	// 查询用户（列表）
	users := oneapi.Users{}
	// 可以模糊查找条件，username LIKE ? or email LIKE ? or display_name LIKE ?", keyword, keyword+"%", keyword+"%", keyword+"%"
	users.Query = map[string]string{
		"keyword": "test1",
	}
	err := users.List(&client)
	// 根据uid获取用户
	user = oneapi.User{}
	for _, u := range users.Users {
        if u.Username == "test1" {
            user.ID = u.ID
        }
    }
	_ = u.Get(&client)
	//更新用户信息
	user.Quota = 500000000
	err = user.Update(&client)
	//删除用户
	//err = user.Delete(&client)

```
```
(2) 渠道（模型）操作
    // 添加渠道
	channel := oneapi.Channel{
		Name:    "ch1",
		BaseUrl: "",
		ChannelConfig: oneapi.ChannelConfig{
			Region: "",
			Sk:     "",
			Ak:     "",
		},
		Group:        "default",
		Models:       "deepseek-r1",
		ModelMapping: "",
		Other:        "",
		SystemPrompt: "",
		Type:         50, //渠道类型是前端constant，参考web/default/src/constants/channel.constants.js 50是openai兼容格式
		Key:          "12345678901234567890",
	}
	err = channel.Add(&client)
	// 查询渠道
	channels := oneapi.Channels{}
	err = channels.List(&client)
	// 可模糊查询
	channels.Query = map[string]string{
		"keyword": "ch1",
	}
	// 修改渠道
	updateChannel := oneapi.Channel{}
	for _, c := range channels.Channels {
		if c.Name == "ch1" {
			updateChannel = *c
		}
	}
	// update channel
	updateChannel.Name = "ch1-updated"
	err = updateChannel.Update(&client)
	//删除渠道
	//err = updateChannel.Delete(&client)
	if err != nil {
		log.Fatal(err)
	}
```
```
 (3) 令牌操作
	// 添加令牌 // expired_time : -1 ,models : "/data/DeepSeek-R1,ERNIE-3.5-8K" ,name : "test" ,remain_quota : 5000000000 ,subnet : "" ,unlimited_quota : false
	token := oneapi.Token{
		Name:           "test1",
		UserID:         user.ID,
		Models:         "/data/DeepSeek-R1,ERNIE-3.5-8K",
		RemainQuota:    5000000000,
		UnlimitedQuota: false,
		ExpiredTime:    -1,
		Subnet:         "",
	}
	err := token.Add(&client)
	// list tokens
	tokens := oneapi.Tokens{}
	// 可模糊查询
	tokens.Query = map[string]string{
        "keyword": "test1",
    }
    // 根据uid获取令牌
    tokens.userID = user.ID
	err := tokens.List(&client, 0)
	if err != nil {
		log.Fatal(err)
	}
	updateToken = tokens.Tokens[0]
	//更新令牌
	updateToken.Models = "/data/DeepSeek-R1"
	updateToken.RemainQuota = 9009000000
	err = updateToken.Update(&client)
	fmt.Println("update token err=", err, "updateToken=", updateToken)

	//删除令牌
	err = updateToken.Delete(&client)
	fmt.Println("delete token err=", err)
```
```
 (4) 日志操作
    // 获取日志
    logs := oneapi.Logs{}
	logs.Query = map[string]string{
		"username": "test",
	}
	err = logs.Get(&client)
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range logs.Logs {
		fmt.Println("l=", *l)
	}
```
