package middleware

import (
	"github.com/gin-gonic/gin"
)

func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "max-age=604800") // one week
		c.Next()
	}
}
