package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"message": "",
		"success": true,
		"data":    data,
	})
}

func Err(c *gin.Context, err error) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": err.Error(),
	})
}

func OpenAiErr(c *gin.Context, err OpenAIError) {
	c.JSON(http.StatusOK, gin.H{
		"error": err,
	})
}
