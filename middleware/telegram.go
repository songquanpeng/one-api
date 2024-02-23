package middleware

import (
	"one-api/common/telegram"
	"os"

	"github.com/gin-gonic/gin"
)

func Telegram() func(c *gin.Context) {
	return func(c *gin.Context) {
		token := c.Param("token")

		if !telegram.TGEnabled || telegram.TGWebHookSecret == "" || token == "" || token != os.Getenv("TG_BOT_API_KEY") {
			c.String(404, "Page not found")
			c.Abort()
			return
		}

		c.Next()
	}
}
