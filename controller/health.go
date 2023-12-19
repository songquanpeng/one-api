package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"one-api/common"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("One-api is healthy. Version: %s", common.Version),
		"data":    "",
	})
	return
}
