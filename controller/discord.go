package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type DiscordOAuthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type DiscordUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func getDiscordUserInfoByCode(codeFromURLParamaters string, host string) (*DiscordUser, error) {
	if codeFromURLParamaters == "" {
		return nil, errors.New("无效参数")
	}

	RequestClient := &http.Client{}

	accessTokenBody := bytes.NewBuffer([]byte(fmt.Sprintf(
		"client_id=%s&client_secret=%s&grant_type=authorization_code&redirect_uri=%s/oauth/discord&code=%s&scope=identify",
		common.DiscordClientId, common.DiscordClientSecret, common.ServerAddress, codeFromURLParamaters,
	)))

	req, _ := http.NewRequest("POST",
		"https://discordapp.com/api/oauth2/token",
		accessTokenBody,
	)

	req.Header = http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
		"Accept":       []string{"application/json"},
	}

	resp, err := RequestClient.Do(req)

	if resp.StatusCode != 200 || err != nil {
		return nil, errors.New("访问令牌无效")
	}

	var discordOAuthResponse DiscordOAuthResponse

	err = json.NewDecoder(resp.Body).Decode(&discordOAuthResponse)

	if err != nil {
		return nil, err
	}

	accessToken := fmt.Sprintf("Bearer %s", discordOAuthResponse.AccessToken)

	// Get User Info
	req, _ = http.NewRequest("GET", "https://discord.com/api/users/@me", nil)

	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{accessToken},
	}

	defer resp.Body.Close()

	resp, err = RequestClient.Do(req)

	if resp.StatusCode != 200 || err != nil {
		return nil, errors.New("Discord 用户信息无效")
	}

	var discordUser DiscordUser

	err = json.NewDecoder(resp.Body).Decode(&discordUser)

	if err != nil {
		return nil, err
	}

	if discordUser.Id == "" {
		return nil, errors.New("返回值无效，用户字段为空，请稍后再试！")
	}

	defer resp.Body.Close()

	return &discordUser, nil
}

func DiscordOAuth(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username != nil {
		DiscordBind(c)
		return
	}

	if !common.DiscordOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Discord 登录以及注册",
		})
		return
	}
	code := c.Query("code")

	discordUser, err := getDiscordUserInfoByCode(code, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		DiscordId: discordUser.Id,
	}
	if model.IsDiscordIdAlreadyTaken(user.DiscordId) {
		err := user.FillUserByDiscordId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		if common.RegisterEnabled {
			user.Username = "discord_" + strconv.Itoa(model.GetMaxUserId()+1)
			if discordUser.Username != "" {
				user.DisplayName = discordUser.Username
			} else {
				user.DisplayName = "Discord User"
			}
			user.Role = common.RoleCommonUser
			user.Status = common.UserStatusEnabled

			if err := user.Insert(0); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
	}

	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

func DiscordBind(c *gin.Context) {
	if !common.DiscordOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Discord 登录以及注册",
		})
		return
	}
	code := c.Query("code")

	discordUser, err := getDiscordUserInfoByCode(code, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		DiscordId: discordUser.Id,
	}
	if model.IsDiscordIdAlreadyTaken(user.DiscordId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Discord 账户已被绑定",
		})
		return
	}
	session := sessions.Default(c)
	id := session.Get("id")
	// id := c.GetInt("id")  // critical bug!
	user.Id = id.(int)
	err = user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.DiscordId = discordUser.Id
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "bind",
	})
}
