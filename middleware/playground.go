package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/playground"
)

// PlaygroundEnabled 游乐场开关中间件：未配置开启时返回 503
func PlaygroundEnabled() func(c *gin.Context) {
	return func(c *gin.Context) {
		if !playground.GetConfig().Enabled {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "模型游乐场未开启，请联系管理员在系统设置中开启",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
