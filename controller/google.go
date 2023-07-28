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

type GoogleAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type GoogleUser struct {
	Sub  string `json:"sub"`
	Name string `json:"name"`
}

func getGoogleUserInfoByCode(codeFromURLParamaters string, host string) (*GoogleUser, error) {
	if codeFromURLParamaters == "" {
		return nil, errors.New("无效参数")
	}

	RequestClient := &http.Client{}

	accessTokenBody := bytes.NewBuffer([]byte(fmt.Sprintf(
		"code=%s&client_id=%s&client_secret=%s&redirect_uri=%s/oauth/google&grant_type=authorization_code",
		codeFromURLParamaters, common.GoogleClientId, common.GoogleClientSecret, common.ServerAddress,
	)))

	req, _ := http.NewRequest("POST",
		"https://oauth2.googleapis.com/token",
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

	var googleTokenResponse GoogleAccessTokenResponse

	json.NewDecoder(resp.Body).Decode(&googleTokenResponse)

	accessToken := "Bearer " + googleTokenResponse.AccessToken

	// Get User Info
	req, _ = http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)

	req.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{accessToken},
	}

	defer resp.Body.Close()

	resp, err = RequestClient.Do(req)

	if resp.StatusCode != 200 || err != nil {
		return nil, errors.New("Google 用户信息无效")
	}

	var googleUser GoogleUser

	// Parse json to googleUser
	err = json.NewDecoder(resp.Body).Decode(&googleUser)

	if err != nil {
		return nil, err
	}

	if googleUser.Sub == "" {
		return nil, errors.New("返回值无效，用户字段为空，请稍后再试！")
	}

	defer resp.Body.Close()

	return &googleUser, nil
}

func GoogleOAuth(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username != nil {
		GoogleBind(c)
		return
	}

	if !common.GoogleOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Google 登录以及注册",
		})
		return
	}
	code := c.Query("code")

	googleUser, err := getGoogleUserInfoByCode(code, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		GoogleId: googleUser.Sub,
	}
	if model.IsGoogleIdAlreadyTaken(user.GoogleId) {
		err := user.FillUserByGoogleId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		if common.RegisterEnabled {
			user.Username = "google_" + strconv.Itoa(model.GetMaxUserId()+1)
			if googleUser.Name != "" {
				user.DisplayName = googleUser.Name
			} else {
				user.DisplayName = "Google User"
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

func GoogleBind(c *gin.Context) {
	if !common.GoogleOAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 Google 登录以及注册",
		})
		return
	}
	code := c.Query("code")

	googleUser, err := getGoogleUserInfoByCode(code, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		GoogleId: googleUser.Sub,
	}
	if model.IsGoogleIdAlreadyTaken(user.GoogleId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 Google 账户已被绑定",
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
	user.GoogleId = googleUser.Sub
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
