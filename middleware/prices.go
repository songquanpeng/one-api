package middleware

import "github.com/gin-gonic/gin"

func PricesAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		typeParam := c.Query("type")
		if typeParam == "old" {
			AdminAuth()(c)
		} else {
			c.Next()
		}
	}
}
