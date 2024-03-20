package middleware

import (
	"one-api/common/telegram"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Telegram() func(c *gin.Context) {
	return func(c *gin.Context) {
		token := c.Param("token")

		if !telegram.TGEnabled || telegram.TGWebHookSecret == "" || token == "" || token != viper.GetString("TG_BOT_API_KEY") {
			c.String(404, "Page not found")
			c.Abort()
			return
		}

		c.Next()
	}
}
