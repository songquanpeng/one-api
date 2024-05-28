package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/model"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LarkAppAccessTokenResponse struct {
	Code           int    `json:"code"`
	Msg            string `json:"msg"`
	AppAccessToken string `json:"app_access_token"`
}

type LarkUserAccessTokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type LarkUser struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		OpenID string `json:"open_id"`
		Name   string `json:"name"`
	} `json:"data"`
}

func getLarkAppAccessToken() (string, error) {
	values := map[string]string{
		"app_id":     config.LarkClientId,
		"app_secret": config.LarkClientSecret,
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal/", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.SysLog(err.Error())
		return "", errors.New("无法连接至飞书服务器，请稍后重试！")
	}
	defer res.Body.Close()
	var appAccessTokenResponse LarkAppAccessTokenResponse
	err = json.NewDecoder(res.Body).Decode(&appAccessTokenResponse)
	if err != nil {
		return "", err
	}

	if appAccessTokenResponse.Code != 0 {
		return "", errors.New(appAccessTokenResponse.Msg)
	}
	return appAccessTokenResponse.AppAccessToken, nil

}

func getLarkUserAccessToken(code string) (string, error) {
	appAccessToken, err := getLarkAppAccessToken()
	if err != nil {
		return "", err
	}
	values := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", appAccessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		logger.SysLog(err.Error())
		return "", errors.New("无法连接至飞书服务器，请稍后重试！")
	}
	defer res.Body.Close()
	var larkUserAccessTokenResponse LarkUserAccessTokenResponse
	err = json.NewDecoder(res.Body).Decode(&larkUserAccessTokenResponse)
	if err != nil {
		return "", err
	}
	if larkUserAccessTokenResponse.Code != 0 {
		return "", errors.New(larkUserAccessTokenResponse.Msg)
	}
	return larkUserAccessTokenResponse.Data.AccessToken, nil
}

func getLarkUserInfoByCode(code string) (*LarkUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}

	userAccessToken, err := getLarkUserAccessToken(code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", "https://open.feishu.cn/open-apis/authen/v1/user_info", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userAccessToken))
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res2, err := client.Do(req)
	if err != nil {
		logger.SysLog(err.Error())
		return nil, errors.New("无法连接至飞书服务器，请稍后重试！")
	}
	var larkUser LarkUser
	err = json.NewDecoder(res2.Body).Decode(&larkUser)
	if err != nil {
		return nil, err
	}
	return &larkUser, nil
}

func LarkOAuth(c *gin.Context) {
	if !config.LarkAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员未开启通过飞书登录以及注册",
			"success": false,
		})
		return
	}
	session := sessions.Default(c)
	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state is empty or not same",
		})
		return
	}
	username := session.Get("username")
	if username != nil {
		LarkBind(c)
		return
	}
	code := c.Query("code")
	larkUser, err := getLarkUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		LarkId: larkUser.Data.OpenID,
	}
	if model.IsLarkIdAlreadyTaken(user.LarkId) {
		err := user.FillUserByLarkId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		if config.RegisterEnabled {
			user.Username = "lark_" + strconv.Itoa(model.GetMaxUserId()+1)
			if larkUser.Data.Name != "" {
				user.DisplayName = larkUser.Data.Name
			} else {
				user.DisplayName = "Lark User"
			}
			user.Role = config.RoleCommonUser
			user.Status = config.UserStatusEnabled

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

	if user.Status != config.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

func LarkBind(c *gin.Context) {
	if !config.LarkAuthEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "管理员未开启通过飞书登录以及注册",
			"success": false,
		})
		return
	}
	code := c.Query("code")
	larkUser, err := getLarkUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		LarkId: larkUser.Data.OpenID,
	}
	if model.IsLarkIdAlreadyTaken(user.LarkId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该飞书账户已被绑定",
		})
		return
	}
	session := sessions.Default(c)
	id := session.Get("id")
	user.Id = id.(int)
	err = user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.LarkId = larkUser.Data.OpenID
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
