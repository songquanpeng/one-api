package middleware

import (
	"one-api/model"

	"github.com/gin-gonic/gin"
)

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set("group", userGroup)

		c.Next()
	}
}
