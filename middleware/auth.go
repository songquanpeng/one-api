package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
	"one-api/model"
)

func authHelper(c *gin.Context, minRole int) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	id := session.Get("id")
	status := session.Get("status")
	authByToken := false
	if username == nil {
		// Check token
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权进行此操作，未登录或 token 无效",
			})
			c.Abort()
			return
		}
		user := model.ValidateUserToken(token)
		if user != nil && user.Username != "" {
			// Token is valid
			username = user.Username
			role = user.Role
			id = user.Id
			status = user.Status
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "无权进行此操作，token 无效",
			})
			c.Abort()
			return
		}
		authByToken = true
	}
	if status.(int) == common.UserStatusDisabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户已被封禁",
		})
		c.Abort()
		return
	}
	if role.(int) < minRole {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权进行此操作，权限不足",
		})
		c.Abort()
		return
	}
	c.Set("username", username)
	c.Set("role", role)
	c.Set("id", id)
	c.Set("authByToken", authByToken)
	c.Next()
}

func UserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleCommonUser)
	}
}

func AdminAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleAdminUser)
	}
}

func RootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, common.RoleRootUser)
	}
}

// NoTokenAuth You should always use this after normal auth middlewares.
func NoTokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authByToken := c.GetBool("authByToken")
		if authByToken {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "本接口不支持使用 token 进行验证",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// TokenOnlyAuth You should always use this after normal auth middlewares.
func TokenOnlyAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authByToken := c.GetBool("authByToken")
		if !authByToken {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "本接口仅支持使用 token 进行验证",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
