package controller

import (
	"github.com/gin-gonic/gin"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"net/http"
)

func GetGroups(c *gin.Context) {
	groupNames := make([]string, 0)
	for groupName := range billingratio.GroupRatio {
		groupNames = append(groupNames, groupName)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    groupNames,
	})
}
