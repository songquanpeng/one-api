package middleware

import (
	"one-api/common"
	"one-api/common/utils"

	"github.com/gin-gonic/gin"
)

func abortWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": utils.MessageWithRequestId(message, c.GetString(common.RequestIdKey)),
			"type":    "one_api_error",
		},
	})
	c.Abort()
	common.LogError(c.Request.Context(), message)
}
