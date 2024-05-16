package controller

import (
	"net/http"
	"one-api/model"
	"one-api/providers"
	providersBase "one-api/providers/base"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetModelList(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	keys := strings.Split(channel.Key, "\n")
	channel.Key = keys[0]

	provider := providers.GetProvider(&channel, c)
	if provider == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "provider not found",
		})
		return
	}

	modelProvider, ok := provider.(providersBase.ModelListInterface)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "channel not implemented",
		})
		return
	}

	modelList, err := modelProvider.GetModelList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    modelList,
	})
}
