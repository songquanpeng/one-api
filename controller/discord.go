package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	disgoauth "github.com/realTristan/disgoauth"
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
		return nil, errors.New("Invalid parameter")
	}

	// Establish a new discord client
	var dc *disgoauth.Client = disgoauth.Init(&disgoauth.Client{
		ClientID:     common.DiscordClientId,
		ClientSecret: common.DiscordClientSecret,
		RedirectURI:  fmt.Sprintf("https://%s/oauth/discord", host),
		Scopes:       []string{disgoauth.ScopeIdentify, disgoauth.ScopeEmail, disgoauth.ScopeGuilds, disgoauth.ScopeGuildsJoin},
	})

	accessToken, _ := dc.GetOnlyAccessToken(codeFromURLParamaters)

	// Get the authorized user's data using the above accessToken
	userData, _ := disgoauth.GetUserData(accessToken)

	// Create a new DiscordUser
	var discordUser DiscordUser

	// Decode the userData map[string]interface{} into the discordUser
	// Convert the map to JSON
	jsonData, _ := json.Marshal(userData)

	// Convert the JSON to a struct
	err := json.Unmarshal(jsonData, &discordUser)

	if err != nil {
		return nil, err
	}

	// Add guild member.
	if common.DiscordGuildId != "" && discordUser.Id != "" && common.DiscordBotToken != "" && common.DiscordAllowJoiningGuild == "true" {
		url := fmt.Sprintf("https://discord.com/api/guilds/%s/members/%s", common.DiscordGuildId, discordUser.Id)

		// Set JSON
		map1 := map[string]interface{}{
			// accessToken remove "Bearer "
			"access_token": string(accessToken[7:]),
		}

		// Convert map to JSON
		jsonData, _ := json.Marshal(map1)

		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

		// Set Header
		req.Header.Set("Authorization", fmt.Sprintf("Bot %s", common.DiscordBotToken))
		req.Header.Set("Content-Type", "application/json")

		// Create a new HTTP Client
		client := &http.Client{}
		resp, err := client.Do(req)

		log.Print(resp.StatusCode)

		if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 201) {
			// Print content
			stringBuff := new(bytes.Buffer)
			stringBuff.ReadFrom(resp.Body)

			// Print error
			fmt.Println("Error: ", stringBuff.String())

			return nil, errors.New("You must join the discord server first or be verified member to be able to login!")
		}

		// Close the response body
		defer resp.Body.Close()
	}

	if discordUser.Username == "" {
		return nil, errors.New("Invalid return value, user field is empty, please try again later!")
	}

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
	host := c.Request.Host
	discordUser, err := getDiscordUserInfoByCode(code, host)
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
	return
}
